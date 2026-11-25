package agentbrowser

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ahobsonsayers/browserfull/internal/config"
	"github.com/tidwall/gjson"
)

const executableName = "agent-browser"

const defaultConfigContents = `{
  "$schema": "https://agent-browser.dev/schema.json"
}`

type Client struct {
	cfg      *config.Config
	execPath string
}

type SessionInfo struct {
	SessionName string
	CDPURL      string
	PID         int
	Engine      string
	StreamPort  int
	Version     string
}

func New(cfg *config.Config) (*Client, error) {
	execPath, err := exec.LookPath(executableName)
	if err != nil {
		return nil, fmt.Errorf("%s executable not found on path: %w", executableName, err)
	}

	return &Client{
		cfg:      cfg,
		execPath: execPath,
	}, nil
}

type LaunchOptions struct {
	Headed bool
}

func (c *Client) Launch(sessionName string, opts LaunchOptions) (*SessionInfo, error) {
	// A headed browser needs an X display; without one Chrome refuses to start
	if opts.Headed && os.Getenv("DISPLAY") == "" {
		log.Print("headed requested but DISPLAY is not set, launching headless")
		opts.Headed = false
	}

	args := []string{"open", "about:blank", "--session", sessionName, "--json"}
	if opts.Headed {
		args = append(args, "--headed")
	}

	_, err := c.runCmd(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	return c.GetSession(sessionName)
}

func (c *Client) Close(sessionName string) error {
	_, err := c.runCmd(
		"close",
		"--session", sessionName,
		"--json",
	)
	if err != nil {
		return fmt.Errorf("failed to close browser: %w", err)
	}

	return nil
}

func (c *Client) ListSessions() ([]string, error) {
	output, err := c.runCmd("session", "list", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get browser sessions: %w", err)
	}

	sessionsResult := gjson.GetBytes(output, "data.sessions")
	if !sessionsResult.IsArray() {
		return nil, fmt.Errorf("sessions not found in output: %s", strings.TrimSpace(string(output)))
	}

	sessionResults := sessionsResult.Array()
	sessions := make([]string, 0, len(sessionResults))
	for _, sessionResult := range sessionResults {
		sessions = append(sessions, sessionResult.String())
	}

	return sessions, nil
}

func (c *Client) GetSession(sessionName string) (*SessionInfo, error) {
	baseDir := c.cfg.DataDir

	pidFile := filepath.Join(baseDir, sessionName+".pid")
	engineFile := filepath.Join(baseDir, sessionName+".engine")
	streamFile := filepath.Join(baseDir, sessionName+".stream")
	versionFile := filepath.Join(baseDir, sessionName+".version")

	pid, err := readIntFile(pidFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pid: %w", err)
	}

	engine, err := readFile(engineFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read engine: %w", err)
	}

	streamPort, err := readIntFile(streamFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream port: %w", err)
	}

	version, err := readFile(versionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	cdpURL, err := c.getCDPURL(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cdp url: %w", err)
	}

	return &SessionInfo{
		SessionName: sessionName,
		CDPURL:      cdpURL,
		PID:         pid,
		Engine:      engine,
		StreamPort:  streamPort,
		Version:     version,
	}, nil
}

// runCmd executes an agent-browser command with the configured env and
// returns its stdout output. It ensures the config file exists first.
func (c *Client) runCmd(args ...string) ([]byte, error) {
	err := c.ensureConfigFile()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(c.execPath, args...)
	env := append(
		os.Environ(),
		fmt.Sprintf("AGENT_BROWSER_CONFIG=%s", c.cfg.ConfigFilePath()),
		fmt.Sprintf("AGENT_BROWSER_SOCKET_DIR=%s", c.cfg.DataDir),
	)
	if c.cfg.BrowserExecPath != "" {
		env = append(env, fmt.Sprintf("AGENT_BROWSER_EXECUTABLE_PATH=%s", c.cfg.BrowserExecPath))
	}
	cmd.Env = env
	return cmd.CombinedOutput()
}

// ensureConfigFile ensures an agent-browser config file exists in the
// data dir, creating it with the default content if it does not
func (c *Client) ensureConfigFile() error {
	configFilePath := c.cfg.ConfigFilePath()

	_, err := os.Stat(configFilePath)
	if err == nil {
		return nil // File exists
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("config file '%s' has an issue: %w", configFilePath, err)
	}

	err = os.MkdirAll(c.cfg.DataDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create data dir '%s' : %w", c.cfg.DataDir, err)
	}

	err = os.WriteFile(c.cfg.ConfigFilePath(), []byte(defaultConfigContents), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write default config to '%s': %w", configFilePath, err)
	}

	return nil
}

func (c *Client) getCDPURL(sessionName string) (string, error) {
	output, err := c.runCmd(
		"get", "cdp-url",
		"--session", sessionName,
		"--json",
	)
	if err != nil {
		return "", fmt.Errorf("failed to get cdp url: %w", err)
	}

	cdpURL := gjson.GetBytes(output, "data.cdpUrl").String()
	if cdpURL == "" {
		return "", fmt.Errorf("cdp url not found in output: %s", strings.TrimSpace(string(output)))
	}

	return cdpURL, nil
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", path, err)
	}

	return strings.TrimSpace(string(content)), nil
}

func readIntFile(path string) (int, error) {
	content, err := readFile(path)
	if err != nil {
		return 0, err
	}

	intContent, err := strconv.Atoi(content)
	if err != nil {
		return 0, fmt.Errorf("failed to parse content of file '%s' as an integer: %w", path, err)
	}

	return intContent, nil
}

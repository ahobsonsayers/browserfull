package api

import "context"

func (s Server) ListSessions(
	_ context.Context,
	_ ListSessionsRequestObject,
) (ListSessionsResponseObject, error) {
	sessions, err := s.agentBrowser.ListSessions()
	if err != nil {
		return nil, err
	}

	return ListSessions200JSONResponse(sessions), nil
}

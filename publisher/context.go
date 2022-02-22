package publisher

import (
	"context"
	"fmt"
)

var (
	AgentUUIDContextKey    = "agentUUID"
	PlatformAddrContextKey = "platformAddr"
	StateContextKey        = "agentState"
)

func stateFromContext(ctx context.Context) (*AgentState, error) {
	key := StateContextKey
	v, ok := ctx.Value(key).(*AgentState)
	if !ok {
		return nil, fmt.Errorf("cannot get %s from ctx %v", AgentUUIDContextKey, ctx)
	}

	return v, nil
}

func stringFromContext(ctx context.Context, key string) (string, error) {
	v, ok := ctx.Value(key).(string)
	if !ok {
		return "", fmt.Errorf("cannot get %s from ctx %v", AgentUUIDContextKey, ctx)
	}

	return v, nil
}

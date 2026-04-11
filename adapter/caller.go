package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"time"
)

// DefaultTimeout is the max time an adapter has to respond.
const DefaultTimeout = 500 * time.Millisecond

// Call invokes the adapter binary with "suggest" subcommand, piping the
// envelope as JSON to stdin and reading a SuggestResponse from stdout.
// Returns nil on timeout, non-zero exit, or decode error.
func Call(adapterPath string, envelope ContextEnvelope) (*SuggestResponse, error) {
	return CallWithTimeout(adapterPath, envelope, DefaultTimeout)
}

// CallWithTimeout is like Call but accepts a custom timeout.
func CallWithTimeout(adapterPath string, envelope ContextEnvelope, timeout time.Duration) (*SuggestResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	input, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, adapterPath, "suggest")
	cmd.Stdin = bytes.NewReader(input)

	out, err := cmd.Output()
	if err != nil {
		// timeout, non-zero exit, signal — all return nil
		return nil, nil //nolint:nilerr
	}

	var resp SuggestResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, nil //nolint:nilerr
	}

	return &resp, nil
}

// CallDescribe invokes the adapter binary with "describe" subcommand,
// reading a DescribeResponse from stdout. Returns nil on error.
func CallDescribe(adapterPath string) (*DescribeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, adapterPath, "describe")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil //nolint:nilerr
	}

	var resp DescribeResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, nil //nolint:nilerr
	}

	return &resp, nil
}

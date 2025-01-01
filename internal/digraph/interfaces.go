// Copyright (C) 2025 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package digraph

import "context"

// Finder finds a DAG by name.
type Finder interface {
	FindByName(ctx context.Context, name string) (*DAG, error)
}

// HistoryStoreClient gets a result of a DAG execution.
type HistoryStoreClient interface {
	GetStatus(ctx context.Context, name string, requestID string) (*HistoryStatus, error)
}

// HistoryStatus is the result of a DAG execution.
type HistoryStatus struct {
	// Name represents the name of the executed DAG.
	Name string `json:"name,omitempty"`
	// Params is the parameters of the DAG execution
	Params string `json:"params,omitempty"`
	// Outputs is the outputs of the DAG execution.
	Outputs map[string]string `json:"outputs,omitempty"`
}
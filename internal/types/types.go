// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.

package types

import "encoding/json"

type CreateTaskRequest struct {
	Input json.RawMessage `json:"input"`
	Model string          `json:"model"`
}

type CreateTaskResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

type CreditResponse struct {
	Code int     `json:"code"`
	Data float64 `json:"data"`
}

type DownloadUrlResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

type RecordInfoResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

// Code generated by jsonrpc2. DO NOT EDIT.
package internal

import (
	"context"
	"encoding/json"
	jsonrpc2 "github.com/reddec/jsonrpc2"
	shared "tinc-web-boot/web/shared"
)

func RegisterTincWebUI(router *jsonrpc2.Router, wrap shared.TincWebUI) []string {
	router.RegisterFunc("TincWebUI.IssueAccessToken", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 uint `json:"validDays"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		return wrap.IssueAccessToken(ctx, args.Arg0)
	})

	router.RegisterFunc("TincWebUI.Notify", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 string `json:"title"`
			Arg1 string `json:"message"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0, &args.Arg1)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		return wrap.Notify(ctx, args.Arg0, args.Arg1)
	})

	router.RegisterFunc("TincWebUI.Endpoints", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct{}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		return wrap.Endpoints(ctx)
	})

	router.RegisterFunc("TincWebUI.Configuration", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct{}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		return wrap.Configuration(ctx)
	})

	return []string{"TincWebUI.IssueAccessToken", "TincWebUI.Notify", "TincWebUI.Endpoints", "TincWebUI.Configuration"}
}

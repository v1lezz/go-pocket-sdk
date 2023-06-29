package go_pocket_sdk

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_GetAuthorizationURL(t *testing.T) {
	type args struct {
		requestToken string
		redirectUrl  string
	}
	want := func(args args) string {
		return fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", args.requestToken, args.redirectUrl)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				requestToken: "qwe-rty-123",
				redirectUrl:  "http://localhost:80/",
			},
			wantErr: false,
		},
		{
			name: "Empty token",
			args: args{
				requestToken: "",
				redirectUrl:  "http://localhost:80/",
			},
			wantErr: true,
		},
		{
			name: "Empty URL",
			args: args{
				requestToken: "qwe-rty-123",
				redirectUrl:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			gor, err := c.GetAuthorizationURL(tt.args.requestToken, tt.args.redirectUrl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, want(tt.args), gor)
			}
		})
	}
}

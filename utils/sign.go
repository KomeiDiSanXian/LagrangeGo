package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	signLogger = GetLogger("sign")

	SIGN_PKG_LIST sync.Map
)

func init() {
	signPkgList := []string{
		"trpc.o3.ecdh_access.EcdhAccess.SsoEstablishShareKey",
		"trpc.o3.ecdh_access.EcdhAccess.SsoSecureAccess",
		"trpc.o3.report.Report.SsoReport",
		"MessageSvc.PbSendMsg",
		// "wtlogin.trans_emp",
		"wtlogin.login",
		// "trpc.login.ecdh.EcdhService.SsoKeyExchange",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginNewDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLoginUnusualDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginUnusualDevice",
		"OidbSvcTrpcTcp.0x11ec_1",
		"OidbSvcTrpcTcp.0x758_1",
		"OidbSvcTrpcTcp.0x7c2_5",
		"OidbSvcTrpcTcp.0x10db_1",
		"OidbSvcTrpcTcp.0x8a1_7",
		"OidbSvcTrpcTcp.0x89a_0",
		"OidbSvcTrpcTcp.0x89a_15",
		"OidbSvcTrpcTcp.0x88d_0",
		"OidbSvcTrpcTcp.0x88d_14",
		"OidbSvcTrpcTcp.0x112a_1",
		"OidbSvcTrpcTcp.0x587_74",
		"OidbSvcTrpcTcp.0x1100_1",
		"OidbSvcTrpcTcp.0x1102_1",
		"OidbSvcTrpcTcp.0x1103_1",
		"OidbSvcTrpcTcp.0x1107_1",
		"OidbSvcTrpcTcp.0x1105_1",
		"OidbSvcTrpcTcp.0xf88_1",
		"OidbSvcTrpcTcp.0xf89_1",
		"OidbSvcTrpcTcp.0xf57_1",
		"OidbSvcTrpcTcp.0xf57_106",
		"OidbSvcTrpcTcp.0xf57_9",
		"OidbSvcTrpcTcp.0xf55_1",
		"OidbSvcTrpcTcp.0xf67_1",
		"OidbSvcTrpcTcp.0xf67_5",
	}

	for _, cmd := range signPkgList {
		SIGN_PKG_LIST.Store(cmd, true)
	}
}

func containSignPKG(cmd string) bool {
	_, ok := SIGN_PKG_LIST.Load(cmd)
	return ok
}

func SignProvider(rawUrl string) func(string, int, []byte) map[string]string {
	return func(cmd string, seq int, buf []byte) map[string]string {
		if !containSignPKG(cmd) {
			return nil
		}
		startTime := time.Now().UnixMilli()
		resp := signResponse{}
		err := httpGet(rawUrl, map[string]string{
			"cmd": cmd,
			"seq": strconv.Itoa(seq),
			"src": fmt.Sprintf("%x", buf),
		}, time.Duration(5)*time.Second, &resp)
		if err != nil {
			signLogger.Error(err)
			return nil
		}

		signLogger.Debugf("signed for [%s:%d](%dms)",
			cmd, seq, time.Now().UnixMilli()-startTime)

		return map[string]string{
			"sign":  resp.Value.Sign,
			"extra": resp.Value.Extra,
			"token": resp.Value.Token,
		}
	}
}

func httpGet(rawUrl string, queryParams map[string]string, timeout time.Duration, target interface{}) error {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("request timed out")
		}
		return fmt.Errorf("failed to perform GET request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()

	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}

type signResponse struct {
	Value struct {
		Sign  string `json:"sign"`
		Extra string `json:"extra"`
		Token string `json:"token"`
	} `json:"value"`
}
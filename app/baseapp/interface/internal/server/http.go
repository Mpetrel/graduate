package server

import (
	v1 "base-service/api/baseapp/interface/v1"
	"base-service/app/baseapp/interface/internal/conf"
	"base-service/app/baseapp/interface/internal/pkg/resp"
	"base-service/app/baseapp/interface/internal/service"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	mmd "github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	nHttp "net/http"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, baseapp *service.BaseappInterfaceService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
			metrics.Server(),
			validate.Validator(),
			mmd.Server(mmd.WithPropagatedPrefix(
				"uid",
			)),
		),
		http.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedHeaders([]string{"Content-Type", "AuthToken"}),
			handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS"}),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	opts = append(opts, http.ResponseEncoder(PrimaryResponseEncoder))
	srv := http.NewServer(opts...)
	v1.RegisterBaseappInterfaceHTTPServer(srv, baseapp)
	return srv
}

// PrimaryResponseEncoder 包装返回值，统一返回格式
func PrimaryResponseEncoder(w nHttp.ResponseWriter, r *nHttp.Request, v interface{}) error {
	any, err := anypb.New(v.(proto.Message))
	if err != nil {
		return err
	}
	result := &resp.ApiResult{
		Code:    200,
		Message: "OK",
		Data:    any,
	}
	codec, _ := http.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(result)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/%s", codec.Name()))
	if sc, ok := v.(interface {
		StatusCode() int
	}); ok {
		w.WriteHeader(sc.StatusCode())
	}
	_, _ = w.Write(data)
	return nil
}
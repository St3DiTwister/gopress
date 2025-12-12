package grpc

import (
	"net"

	articlepb "gopress/api/proto/article"
	authpb "gopress/api/proto/auth"
	"gopress/internal/grpc/interceptor"
	"gopress/internal/repository"
	jwtpkg "gopress/pkg/jwt"

	"google.golang.org/grpc"
)

type Server struct {
	srv  *grpc.Server
	lis  net.Listener
	addr string
}

func NewServer(userRepo repository.UserRepo, articleRepo repository.ArticleRepo, jwtManager *jwtpkg.Manager, addr string) (*Server, error) {
	authI := interceptor.NewAuthInterceptor(jwtManager, []string{
		// публичные auth методы:
		"/auth.AuthService/Register",
		"/auth.AuthService/Login",

		// публичные методы статей:
		"/article.ArticleService/List",
		"/article.ArticleService/Get",
	})

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(authI.Unary()),
	)

	// регистрируем сервисы
	authpb.RegisterAuthServiceServer(grpcSrv, NewAuthServer(userRepo, jwtManager))
	articlepb.RegisterArticleServiceServer(grpcSrv, NewArticleServer(articleRepo))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{srv: grpcSrv, lis: lis, addr: addr}, nil
}

func (s *Server) Start() error {
	return s.srv.Serve(s.lis)
}

func (s *Server) Stop() {
	s.srv.GracefulStop()
}

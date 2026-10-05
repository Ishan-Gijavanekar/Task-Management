package interceptor

var PublicMethods = []string{
	"/auth.v1.AuthService/Register",
	"/auth.v1.AuthService/Login",

	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

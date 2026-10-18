package tests

import (
	"testing"

	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	authv1 "github.com/Ishan-Gijavanekar/user-service/api/proto/auth/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func requireMethod(t *testing.T, services protoreflect.ServiceDescriptors, serviceName, methodName string) {
	t.Helper()
	service := services.ByName(protoreflect.Name(serviceName))
	if service == nil {
		t.Fatalf("protobuf service %q is missing", serviceName)
	}
	if service.Methods().ByName(protoreflect.Name(methodName)) == nil {
		t.Fatalf("protobuf method %s.%s is missing", serviceName, methodName)
	}
}

func TestPublicGRPCContract(t *testing.T) {
	userServices := userv1.File_api_proto_user_proto.Services()
	for _, method := range []string{"CreateUser", "UpdateUser", "DeleteUser", "GetUser", "GetUsers"} {
		requireMethod(t, userServices, "UserService", method)
	}
	authServices := authv1.File_api_proto_auth_auth_proto.Services()
	for _, method := range []string{"Register", "Login"} {
		requireMethod(t, authServices, "AuthService", method)
	}
}

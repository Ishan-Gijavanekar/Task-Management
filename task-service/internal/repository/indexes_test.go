package repository

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestWorkspaceIsFirstFieldInTaskQueryIndex(t *testing.T) {
	index := bson.D{{Key: "workspaceId", Value: 1}, {Key: "projectId", Value: 1}, {Key: "status", Value: 1}}
	if index[0].Key != "workspaceId" {
		t.Fatal("task indexes must begin with workspaceId")
	}
}

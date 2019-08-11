// +build integration contract

package main

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/xuanit/testing/todo/consumer"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"

	grunting "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pact-foundation/pact-go/types"
	"github.com/xuanit/testing/todo/pb"
	"github.com/xuanit/testing/todo/server/service"
	"google.golang.org/grpc"
)

var grpcAddress = "localhost:5003"
var httpAddress = "localhost:5004"

type toDoImplStub struct {
}

func (r toDoImplStub) List(limit int32, notCompleted bool) ([]*pb.Todo, error) {
	return toDos, nil
}

func (r toDoImplStub) Insert(items *pb.Todo) error {
	return nil
}

func (r toDoImplStub) Get(id string) (*pb.Todo, error) {
	return nil, nil
}

func (r toDoImplStub) Delete(id string) error {
	return nil
}

var toDos []*pb.Todo

func startServer() {
	todoRep := &toDoImplStub{}
	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, service.ToDo{ToDoRepo: todoRep})

	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatal("can not listen tcp grpcAddress ", grpcAddress, " ", err)
	}

	log.Printf("Serving GRPC at %s.\n", grpcAddress)
	go s.Serve(lis)

	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Couldn't contact grpc server")
	}

	mux := grunting.NewServeMux()
	err = pb.RegisterTodoServiceHandler(context.Background(), mux, conn)
	if err != nil {
		panic("Cannot serve http api")
	}
	log.Printf("Serving http at %s.\n", httpAddress)
	err = http.ListenAndServe(httpAddress, mux)
}

func TestToDoService(t *testing.T) {
	var dir, _ = os.Getwd()
	var pactDir = fmt.Sprintf("%s/../consumer/pacts", dir)
	go startServer()

	pact := &dsl.Pact{
		Consumer: "ToDoConsumer",
		Provider: "ToDoService",
	}
	pact.DisableToolValidityCheck = true

	// Verify the Provider using the locally saved Pact Files
	_, _ = pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: "http://" + httpAddress,
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/todoconsumer-todoservice.json", pactDir))},
		StateHandlers: types.StateHandlers{
			// Setup any state required by the test
			// in this case, we ensure there is a "user" in the system
			"There are todo A and todo B": func() error {
				toDos = []*pb.Todo{{Id: "id1", Title: "ToDo A"}, {Id: "id2", Title: "ToDo B"}}
				return nil
			},
		},
	})

}

func TestCreateToDo(t *testing.T) {
	pactDir, pact := getPact()

	defer pact.Teardown()

	_, _ = pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: "http://" + httpAddress,
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/todoconsumer-todoservice.json", pactDir))},
	})

	var test = func() (err error) {
		proxy := consumer.ToDoProxy{Host: "localhost", Port: pact.Server.Port}
		id, err := proxy.CreateToDo(consumer.ToDo{Id: uuid.NewV4().String(), Title: "Another todo", Description: "Another description", Completed: true})
		if err != nil {
			return err
		}
		assert.Equal(t, "id1", id)
		return nil
	}

	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

}

func getPact() (string, *dsl.Pact) {
	var dir, _ = os.Getwd()
	var pactDir = fmt.Sprintf("%s/../consumer/pacts", dir)
	go startServer()
	pact := &dsl.Pact{
		Consumer: "ToDoConsumer",
		Provider: "ToDoService",
	}
	pact.DisableToolValidityCheck = true
	return pactDir, pact
}

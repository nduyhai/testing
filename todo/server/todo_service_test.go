// +build integration contract

package main

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
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
var pactPath = "/opt/pact/bin"

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
				log.Println("'There are todo A and todo B' state handler invoked")
				toDos = []*pb.Todo{{Id: "id1", Title: "ToDo A"}, {Id: "id2", Title: "ToDo B"}}
				return nil
			},
		},
	})

}

func TestCreateToDo(t *testing.T) {
	pact := getPact()

	defer pact.Teardown()

	todoReq := consumer.ToDo{Id: uuid.NewV4().String(), Title: "1-1 with manager", Description: "discuss about OKRs", Completed: true}
	// Pass in test case
	var test = func() (err error) {
		proxy := consumer.ToDoProxy{Host: "localhost", Port: pact.Server.Port}
		_, err = proxy.CreateToDo(todoReq)
		if err != nil {
			return err
		}
		return nil
	}

	// Set up our expected interactions.
	pact.AddInteraction().
		UponReceiving("A request to create todo").
		WithRequest(dsl.Request{
			Method:  "POST",
			Path:    dsl.String("/v1/todo"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body: map[string]interface{}{
				"completed":   todoReq.Completed,
				"description": todoReq.Description,
				"title":       todoReq.Title,
			},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body:    dsl.Like(&consumer.ToDo{Id: todoReq.Id}),
		})

	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

}

func TestListToDo(t *testing.T) {
	pact := getPact()

	defer pact.Teardown()

	// Set up our expected interactions.
	pact.AddInteraction().
		Given("There are todo A and todo B").
		UponReceiving("A request to create todo").
		Given("User foo exists").
		WithRequest(dsl.Request{
			Method:  "GET",
			Path:    dsl.Like("/v1/todo?limit=10&not_completed=true"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body:    dsl.Like(&consumer.ToDoList{}),
		})
	// Pass in test case
	var test = func() (err error) {
		proxy := consumer.ToDoProxy{Host: "localhost", Port: pact.Server.Port}
		doList, err := proxy.ListToDo(10, true)

		if err != nil {
			return err
		}
		log.Println("Result:", doList)
		return nil
	}

	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}
func getPact() *dsl.Pact {
	env := os.Getenv("PATH")
	_ = os.Setenv("PATH", env+":"+pactPath)
	go startServer()
	pact := &dsl.Pact{
		Consumer: "ToDoConsumer",
		Provider: "ToDoService",
		Host:     "localhost",
		LogLevel: "DEBUG",
	}
	pact.DisableToolValidityCheck = true
	return pact
}

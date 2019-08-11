// +build integration persistence

package repository

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/suite"
	"github.com/xuanit/testing/todo/pb"
)

type ToDoRepositorySuite struct {
	db *pg.DB
	suite.Suite
	todoRep ToDoImpl
}

func (s *ToDoRepositorySuite) SetupSuite() {
	// Connect to PostgresQL
	s.db = pg.Connect(&pg.Options{
		User:                  "postgres",
		Password:              "example",
		Database:              "todo",
		Addr:                  "localhost" + ":" + "5433",
		RetryStatementTimeout: true,
		MaxRetries:            4,
		MinRetryBackoff:       250 * time.Millisecond,
	})

	// Create Table
	_ = s.db.CreateTable(&pb.Todo{}, nil)

	s.todoRep = ToDoImpl{DB: s.db}
}

func (s *ToDoRepositorySuite) TearDownSuite() {
	_ = s.db.DropTable(&pb.Todo{}, nil)
	_ = s.db.Close()
}

func (s *ToDoRepositorySuite) TestInsert() {
	item := &pb.Todo{Id: uuid.NewV4().String(), Title: "meeting"}
	err := s.todoRep.Insert(item)

	s.Nil(err)

	newTodo, err := s.todoRep.Get(item.Id)
	s.Nil(err)
	s.Equal(item, newTodo)
}

func (s *ToDoRepositorySuite) TestList() {
	item := &pb.Todo{Id: uuid.NewV4().String(), Title: "meeting"}
	err := s.todoRep.Insert(item)
	s.Nil(err)

	cases := []struct {
		Name      string
		Limit     int32
		Completed bool
		FnResult  func([]*pb.Todo, error)
	}{
		{
			"With completed true",
			10,
			true,
			func(toDos []*pb.Todo, err error) {
				fmt.Println(toDos, err)
				s.Nil(err)
				s.Nil(toDos)
			},
		},
		{
			"Struct with completed false",
			10,
			false,
			func(toDos []*pb.Todo, err error) {
				fmt.Println(toDos, err)
				s.Nil(err)
				s.NotNil(toDos)
			},
		},
	}

	for _, test := range cases {
		s.Suite.Run(test.Name, func() {
			toDos, err := s.todoRep.List(test.Limit, test.Completed)
			test.FnResult(toDos, err)
		})
	}
}

func (s *ToDoRepositorySuite) TestGet() {
	item := &pb.Todo{Id: uuid.NewV4().String(), Title: "meeting"}
	err := s.todoRep.Insert(item)

	s.Nil(err)

	newTodo, err := s.todoRep.Get(item.Id)
	s.Nil(err)
	s.Equal(item, newTodo)
}

func (s *ToDoRepositorySuite) TestDelete() {
	item := &pb.Todo{Id: uuid.NewV4().String(), Title: "meeting"}
	err := s.todoRep.Insert(item)

	s.Nil(err)

	err = s.todoRep.Delete(item.Id)

	s.Nil(err)
}

func TestToDoRepository(t *testing.T) {
	suite.Run(t, new(ToDoRepositorySuite))
}

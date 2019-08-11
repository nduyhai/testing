package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xuanit/testing/todo/pb"
	"github.com/xuanit/testing/todo/server/repository/mocks"
	"testing"
)

func TestToDo_CreateTodo(t *testing.T) {
	mockToDoRep := &mocks.ToDo{}
	req := &pb.CreateTodoRequest{Item: &pb.Todo{}}
	mockToDoRep.On("Insert", mock.Anything).Return(nil)

	service := ToDo{ToDoRepo: mockToDoRep}

	res, err := service.CreateTodo(nil, req)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Id)
	mockToDoRep.AssertExpectations(t)
}

func TestToDo_GetTodo(t *testing.T) {
	mockToDoRep := &mocks.ToDo{}
	toDo := &pb.Todo{}
	req := &pb.GetTodoRequest{Id: "123"}
	mockToDoRep.On("Get", req.Id).Return(toDo, nil)
	service := ToDo{ToDoRepo: mockToDoRep}

	res, err := service.GetTodo(nil, req)

	expectedRes := &pb.GetTodoResponse{Item: toDo}

	assert.Nil(t, err)
	assert.Equal(t, expectedRes, res)
	mockToDoRep.AssertExpectations(t)
}

func TestToDo_ListTodo(t *testing.T) {
	mockToDoRep := &mocks.ToDo{}
	var toDos = make([]*pb.Todo, 1)
	req := &pb.ListTodoRequest{Limit: 2, NotCompleted: true}
	expectedRes := &pb.ListTodoResponse{Items: toDos}

	mockToDoRep.On("List", req.Limit, req.NotCompleted).Return(toDos, nil)

	service := ToDo{ToDoRepo: mockToDoRep}

	res, err := service.ListTodo(nil, req)

	assert.Nil(t, err)
	assert.Equal(t, expectedRes, res)
	mockToDoRep.AssertExpectations(t)
}

func TestToDo_DeleteTodo(t *testing.T) {
	mockToDoRep := &mocks.ToDo{}
	req := &pb.DeleteTodoRequest{Id: "123"}
	expectedRes := &pb.DeleteTodoResponse{}
	mockToDoRep.On("Delete", req.Id).Return(nil)

	service := ToDo{ToDoRepo: mockToDoRep}

	res, err := service.DeleteTodo(nil, req)
	assert.Nil(t, err)
	assert.Equal(t, expectedRes, res)
	mockToDoRep.AssertExpectations(t)

}

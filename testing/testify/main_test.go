package main

import (
	"fmt"
	"github.com/samber/lo"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestObjectImplementation struct {
	mock.Mock
}

type TestObjectSuiteImplementation struct {
	suite.Suite
	startPoint int
}

func printArgumentFunction(str string) {
	fmt.Print(str)
}
func printArgumentLenFunction(str string) {
	fmt.Print(len(str))
}

//func Test_Nothing(t *testing.T) {
//	object := &Object{}
//	Function(object, printArgumentFunction)
//}

func (suite *TestObjectSuiteImplementation) SetupTest() {
	suite.startPoint = 5
}

func Test_Object_F(t *testing.T) {
	var mockedService = new(TestObjectImplementation)
	c := mockedService.On("Method_A").Return("arg")
	assert.Equal(t, []*mock.Call{c}, mockedService.ExpectedCalls)
	assert.Equal(t, "Method_A", c.Method)
	assert.Equal(t, "arg", c.ReturnArguments.Get(0))
}

func Test_Object_F_TestData(t *testing.T) {
	var mockedService = new(TestObjectImplementation)
	mockedService.
		On("Method_B", 0).Return("0", nil)
	expectedCalls := []*mock.Call{
		{
			Parent:          &mockedService.Mock,
			Method:          "Method_B",
			Arguments:       []interface{}{0},
			ReturnArguments: []interface{}{"0", nil},
		},
	}
	assert.Equal(t, "Method_B", expectedCalls[0].Method)
	assert.Equal(t, "0", expectedCalls[0].ReturnArguments[0])
	assert.Equal(t, nil, expectedCalls[0].ReturnArguments[1])
	assert.Equal(t, 0, expectedCalls[0].Arguments[0])
}

func Test_Object_AnythingOfType(t *testing.T) {
	var mockedService = new(TestObjectImplementation)
	c := mockedService.On("Method_C", mock.AnythingOfType("func(int)error)")).Return(0, nil)

	assert.Equal(t, []*mock.Call{c}, mockedService.ExpectedCalls)
	assert.Equal(t, 0, c.ReturnArguments[0])
	assert.Equal(t, nil, c.ReturnArguments[1])
	assert.Equal(t, "Method_C", c.Method)
}

func (suite *TestObjectSuiteImplementation) TestExample() {
	assert.Equal(suite.T(), 5, suite.startPoint)
	suite.Equal(5, suite.startPoint)
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(TestObjectSuiteImplementation))
}

func TestExampleTestSuite2(t *testing.T) {
	is := assert.New(t)

	d := []data{
		{value: 1, name: "aaa"},
		{value: 2, name: "bb0"},
		{value: 0, name: "ccc"},
		{value: 4, name: "ddd"},
		{value: 4, name: "ddd"}}

	r := lo.Filter[data](d, func(x data, _ int) bool {
		return x.value != 0
	})

	r = lo.Uniq[data](r)

	is.Equal(true, true)

	rr := lo.GroupBy[data, int](r, func(i data) int {
		return i.value
	})

	fmt.Println(rr)
	//lo.ForEach()

	sum := lo.Reduce[data, int](r, func(agg int, item data, _ int) int {
		return agg + item.value
	}, 0)

	fmt.Println(sum)

}

type data struct {
	name  string
	value int
}

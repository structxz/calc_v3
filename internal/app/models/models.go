package models

import (
	"time"
)

var (
	StatusPending  string = "PENDING"
	StatusProgress string = "IN_PROGRESS"
	StatusComplete string = "COMPLETE"
	StatusError    string = "ERROR"
)

type Expression struct {
	ID         string    `json:"id"`
	Expression string    `json:"expression,omitempty"`
	Status     string    `json:"status"`
	Result     *float64  `json:"result,omitempty"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	Error      string    `json:"error,omitempty"`
}

type Task struct {
	ID               string    `json:"id"`
	ExpressionID     string    `json:"expression_id"`
	Operation        string    `json:"operation"`
	Arg1             float64   `json:"arg1"`
	Arg2             float64   `json:"arg2"`
	Result           *float64  `json:"result,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DependsOnTaskIDs []string  `json:"depends_on_task_ids,omitempty"`
}

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID string `json:"id"`
}

type TaskResult struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

type ExpressionResponse struct {
	Expression Expression `json:"expression"`
}

type ExpressionsResponse struct {
	Expressions []Expression `json:"expressions"`
}

type TaskResponse struct {
	Task Task `json:"task"`
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	status string `json:"status"`
}

type AuthResponse struct {
	status string `json:"status"`
	token  string `json:"token"`
}

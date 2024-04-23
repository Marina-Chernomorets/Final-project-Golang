package main

import (
 "fmt"
 //"io/ioutil"
 "net/http"
 //"net/url"
 "sync"
 "time"

 "github.com/google/uuid"
)

type Expression struct {
 ID          string
 Expression  string
 Status      string
 Result      int
 CreatedAt   time.Time
 CompletedAt time.Time
}
type User struct {
    Login    string
    Password string
}
var users map[string]User


func registerUserHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    login := r.FormValue("login")
    password := r.FormValue("password")
    if login == "" || password == "" {
        http.Error(w, "Login and password are required", http.StatusBadRequest)
        return
    }

    // Проверяем, что пользователь с таким логином еще не зарегистрирован
    _, exists := users[login]
    if exists {
        http.Error(w, "User already exists", http.StatusConflict)
        return
    }

    users[login] = User{Login: login, Password: password}
    w.WriteHeader(http.StatusOK)
}

func loginUserHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    login := r.FormValue("login")
    password := r.FormValue("password")
    if login == "" || password == "" {
        http.Error(w, "Login and password are required", http.StatusBadRequest)
        return
    }

    user, exists := users[login]
    if !exists || user.Password != password {
        http.Error(w, "Invalid login or password", http.StatusUnauthorized)
        return
    }

    // Генерируем JWT токен
    // Здесь может быть логика генерации и подписи токена
    jwtToken := "your_generated_jwt_token_here"

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "JWT token: %s", jwtToken)
}


var (
 expressions    []*Expression
 expressionsMux sync.Mutex
)

func main() {
    users = make(map[string]User)

    http.HandleFunc("/api/v1/register", registerUserHandler)
    http.HandleFunc("/api/v1/login", loginUserHandler)
    http.HandleFunc("/add", addExpressionHandler)
    http.HandleFunc("/expressions", getExpressionsHandler)
    http.HandleFunc("/expression", getExpressionByIDHandler)

    http.ListenAndServe(":8080", nil)

    //testAPI()
}


func calculateExpression(expression string) int {
 result := 0
 for i := 0; i < len(expression); i += 2 {
  if i == 0 {
   result = int(expression[i] - '0')
  } else {
   num := int(expression[i+1] - '0')
   switch expression[i] {
   case '+':
    result += num
   case '-':
    result -= num
   case '*':
    result *= num
   case '/':
    result /= num
   }
  }
 }

 return result
}

func addExpressionHandler(w http.ResponseWriter, r *http.Request) {
 if r.Method != http.MethodPost {
  http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  return
 }

 expression := r.FormValue("expression")
 if expression == "" {
  http.Error(w, "Expression is required", http.StatusBadRequest)
  return
 }

 expressionsMux.Lock()
 defer expressionsMux.Unlock()

 id := uuid.New().String()
 createdAt := time.Now()

 expressions = append(expressions, &Expression{
  ID:         id,
  Expression: expression,
  Status:     "pending",
  CreatedAt:  createdAt,
 })

 go func() {
  result := calculateExpression(expression)

  expressionsMux.Lock()
  defer expressionsMux.Unlock()

  for _, expr := range expressions {
   if expr.ID == id {
    expr.Status = "completed"
    expr.Result = result
    expr.CompletedAt = time.Now()
    break
   }
  }
 }()

 w.WriteHeader(http.StatusOK)
 fmt.Fprintf(w, "Expression added with ID: %s", id)
}

func getExpressionsHandler(w http.ResponseWriter, r *http.Request) {
 if r.Method != http.MethodGet {
  http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  return
 }

 expressionsMux.Lock()
 defer expressionsMux.Unlock()


   w.Header().Set("Content-Type", "text/plain")
   for _, expr := range expressions {
    fmt.Fprintf(w, "ID: %s, Expression: %s, Status: %s, Result: %d, CreatedAt: %s, CompletedAt: %sn",
     expr.ID, expr.Expression, expr.Status, expr.Result, expr.CreatedAt, expr.CompletedAt)
   }
  }

  func getExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
   if r.Method != http.MethodGet {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
   }

   id := r.FormValue("id")
   if id == "" {
    http.Error(w, "ID is required", http.StatusBadRequest)
    return
   }

   expressionsMux.Lock()
   defer expressionsMux.Unlock()

   var found bool
   for _, expr := range expressions {
    if expr.ID == id {
     w.Header().Set("Content-Type", "text/plain")
     fmt.Fprintf(w, "ID: %s, Expression: %s, Status: %s, Result: %d, CreatedAt: %s, CompletedAt: %sn",
      expr.ID, expr.Expression, expr.Status, expr.Result, expr.CreatedAt, expr.CompletedAt)
     found = true
     break
    }
   }

   if !found {
    http.Error(w, "Expression not found", http.StatusNotFound)
   }
  }

package user

type RegisterView struct {
	Username string
	Email    string
	Error    string
}

type RegisterRequest struct {
	Username string `form:"username" binding:"required,min=3,max=32"`
	Email    string `form:"email"    binding:"required,email"`
	Password string `form:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

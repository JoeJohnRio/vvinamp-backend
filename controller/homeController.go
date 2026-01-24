package controller

import (
	"github.com/gin-gonic/gin"
)

// HomeController handles the home page
type HomeController struct{}

// NewHomeController is a constructor for HomeController
func NewHomeController() Result {
	return Result{
		Controller: &HomeController{},
	}
}

// Register is function to register all controller's endpoint handler
func (h *HomeController) Register(r *gin.Engine) {
	r.GET("/", h.HomePage)
	r.GET("/youtube", h.RedirectToYouTube)
}

// HomePage serves the home page with a button to YouTube
func (h *HomeController) HomePage(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Yak</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f0f0f0;
        }
        .container {
            text-align: center;
        }
        .btn {
            background-color: #ff0000;
            color: white;
            padding: 15px 30px;
            font-size: 18px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
        }
        .btn:hover {
            background-color: #cc0000;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="/youtube" class="btn">If this page is running, i'm prob working on something</a>
    </div>
</body>
</html>
`
	c.Data(200, "text/html; charset=utf-8", []byte(html))
}

// RedirectToYouTube redirects to youtube.com
func (h *HomeController) RedirectToYouTube(c *gin.Context) {
	c.Redirect(302, "https://www.youtube.com")
}

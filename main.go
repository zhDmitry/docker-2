package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB
var err error
var port = flag.String("port", ":1234", "port for app")
var connection = flag.String("c", "--", "db connection string for app")

const uploadPath = "./upload/"
const taskFileName = "task.md"

func connectToDb() *gorm.DB {
	db, err = gorm.Open("mysql", *connection)
	if err != nil {
		log.Fatal(err)
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Task{})
	db.AutoMigrate(&User{})

	return db
}

func getDb() *gorm.DB {
	return db
}

func generateNewInvite(c *gin.Context) {
	task := &Task{Info: c.Query("info"), ID: fmt.Sprint(rand.Int())}

	err := getDb().Create(&task).Error
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database issue"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": task.ID})
	return
}

func listTasks(c *gin.Context) {
	var tasks []Task
	err := getDb().Find(&tasks).Error
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server issue see logs"})
	}
	c.JSON(http.StatusOK, tasks)
}

func createDirIfNotExist(dir string) (err error) {
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	}
	return
}

func finishTask(c *gin.Context) {
	inviteID := c.Param("inviteID")
	var task Task
	link := getDb().First(&task, inviteID)
	err := link.Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you haven't task yet, contact admin to get invite"})
	}
	if task.CreatedAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you haven't started task yet, go to /invite/:yourId to start task "})
	}
	timeFinished := time.Now()
	err = link.Updates(&Task{FinishedAt: timeFinished}).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to finish your task connect to admin and send your test task by email"})
		return
	}
	file, _ := c.FormFile("file")
	if file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't handle your file"})
		return

	}

	err = createDirIfNotExist(uploadPath + inviteID)
	if err != nil {
		log.Fatal(err)
	}
	savePath := uploadPath + inviteID + "/" + timeFinished.Format("2006_01_02_15_04") + "_" + file.Filename
	log.Print(task)
	err = c.SaveUploadedFile(file, savePath)
	log.Print(err, inviteID)
	c.JSON(http.StatusOK, gin.H{"success": "hooray you task was successfully submitted"})

}

func startTask(c *gin.Context) {
	var task Task
	link := getDb().First(&task, c.Param("inviteID"))
	err := link.Error
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to start your task contact with admin"})
		return
	}
	log.Print(task)
	var updateError error
	if task.StartedAt.IsZero() {
		updateError = link.Updates(&Task{StartedAt: time.Now()}).Error
	}
	if updateError == nil {
		log.Print(err)
		c.Header("Content-Disposition", "attachment; filename="+taskFileName)
		c.File("./" + taskFileName)

		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "something weird happened"})
	return
}

func taskView(c *gin.Context) {
	var task Task
	err := getDb().First(&task, c.Param("inviteID")).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please contact with admin to get invite"})
		return
	}
	c.HTML(http.StatusOK, "taskview.tmpl", task)
	return
}

func adminApp(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_app.html", gin.H{})
	return
}
func finishView(c *gin.Context) {
	var task Task
	err := getDb().First(&task, c.Param("inviteID")).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please contact with admin to get invite"})
		return
	}
	c.HTML(http.StatusOK, "finish_task_view.tmpl", task)
	return
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	fmt.Println(*connection)
	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Delims("{{{", "}}}")

	r.LoadHTMLGlob("templates/*")

	db := connectToDb()
	defer db.Close()
	authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		"vova":   "1werty",
		"cristy": "Q2erty",
		"radu":   "Qw3rty",
		"dima":   "Qwe4ty",
	}))

	authorized.StaticFS("api/submits", http.Dir("./upload"))
	authorized.GET("api/newinvite", generateNewInvite)
	authorized.GET("api/tasks", listTasks)
	authorized.Any("/", adminApp)

	r.GET("/invite/:inviteID", taskView)
	r.GET("/start/:inviteID", startTask)
	r.POST("/task-finish/:inviteID", finishTask)
	r.GET("/finish/:inviteID", finishView)

	r.Run(*port)
}

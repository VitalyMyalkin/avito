package main

import (
	"context"
	"encoding/json"
	"database/sql"
	"net/http"
	"fmt"

	"github.com/blockloop/scan"
	"github.com/gin-gonic/gin"

	"github.com/VitalyMyalkin/avito/cmd/config"
)

type App struct {
	Cfg     config.Config
	PostgresDB *sql.DB		
}

type Segment struct {
	ID 	  int    `json:"id"`
	Slug  string `json:"slug"`
}

type Request struct {
	Add    []string `json:"add,omitempty"`
	Delete []string `json:"delete,omitempty"`
}

type User struct {  
	UserID 	 string `json:"user_id"`
}

func NewApp() *App {
	cfg := config.ConfigSetup()
	db, err := sql.Open("pgx", cfg)
    if err != nil {
        fmt.Println(err)
    }
    defer db.Close()
	return &App{
		Cfg:     cfg,
		PostgresDB: db,
	}
}

func (newApp *App) AddSegment(c *gin.Context) {
	// создаем экземпляр сегмента
	var newSegment Segment
	// распарсиваем данные нового 
	newSegment.Slug = c.Param("slug")
	//создаем таблицу сегментов, если ее нет
	_, err := newApp.PostgresDB.Exec("CREATE TABLE IF NOT EXISTS segments (id SERIAL PRIMARY KEY, slug TEXT)")
	if err != nil {
		fmt.Println(err)
	}
	//добавляем новый сегмент в таблицу
	err = newApp.PostgresDB.QueryRow("INSERT INTO segments (slug) VALUES ($1) RETURNING id", newSegment.Slug).Scan(&newSegment.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	//отправляем айдишник добавленного сегмента
	c.JSON(http.StatusCreated, gin.H{
		"id": newSegment.ID,
		"slug": newSegment.Slug,
	})
}

func (newApp *App) RemoveSegment(c *gin.Context) {
	// создаем экземпляр сегмента
	var newSegment Segment
	// распарсиваем данные для удаления
	newSegment.Slug = c.Param("slug")
	//убираем сегмент из таблицы
	err := newApp.PostgresDB.QueryRow("DELETE FROM segments WHERE slug = $1", newSegment.Slug)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	//отправляем название удаленного сегмента
	c.JSON(http.StatusOK, gin.H{
		"slug": newSegment.Slug,
	})
}

func (newApp *App) RefreshUserSegments(c *gin.Context) {
	// создаем экземпляр пользователя
	var newUser User
	// распарсиваем данные нового пользователя
	newUser.UserID = c.Param("id")
	//создаем таблицу пользователей и таблицу связей
	_, err := newApp.PostgresDB.Exec("CREATE TABLE IF NOT EXISTS users (id TEXT)")
	if err != nil {
		fmt.Println(err)
	}
	_, err = newApp.PostgresDB.Exec("CREATE TABLE IF NOT EXISTS usersegments (usersID TEXT NOT NULL, segmentsID INT NOT NULL, CONSTRAINT pk_usersegments PRIMARY KEY (usersID, segmentsID), CONSTRAINT fk_usersegments_usersID FOREIGN KEY (usersID) REFERENCES users (id) ON DELETE CASCADE, CONSTRAINT fk_usersegments_segmentsID FOREIGN KEY (segmentsID) REFERENCES segments (id) ON DELETE CASCADE)"	)
	if err != nil {
		fmt.Println(err)
	}

	// распарсиваем сегменты
	var req Request
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	//добавляем юезра в сегменты
	for _, item := range req.Add {
		var newSegment Segment
		newSegment.Slug = item
		err = newApp.PostgresDB.QueryRow(`SELECT id FROM segments WHERE slug = $1`, item).Scan(&newSegment.ID)
		_, err = newApp.PostgresDB.ExecContext(context.Background(), `INSERT INTO usersegments (usersID, segmentsID) VALUES ($1, $2)`, newUser.UserID, newSegment.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
		}
	}

	//удаляем юзера из сегментов 
	for _, item := range req.Delete {
		var newSegment Segment
		newSegment.Slug = item
		err = newApp.PostgresDB.QueryRow(`SELECT id FROM segments WHERE slug = $1`, item).Scan(&newSegment.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
		}
		_, err = newApp.PostgresDB.ExecContext(context.Background(), `DELETE FROM usersegments WHERE usersID = $1 AND segmentsID = $2`, newUser.UserID, newSegment.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
		}
	}
	
	//отправляем, что изменилось
	c.JSON(http.StatusOK, gin.H{
		"user_id": newUser.UserID,
		"segments_added": req.Add,
		"segments_deleted": req.Delete,
	})
}

func (newApp *App) GetUserSegments(c *gin.Context) {
	// создаем экземпляр сегмента
	var segmentsList []Segment
	// создаем экземпляр пользователя
	var newUser User
	// распарсиваем данные нового пользователя
	newUser.UserID = c.Param("id")
	//делаем список сегментов новый сегмент в таблицу
	rows, err := newApp.PostgresDB.Query("SELECT * FROM segments WHERE id = (SELECT segmentsID FROM usersegments WHERE usersID = $1)", newUser.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	err = scan.Rows(&segmentsList, rows)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	
	//отправляем сегменты, в которых есть юзер
	c.JSON(http.StatusOK, gin.H{
		"user_id": newUser.UserID,
		"segments": segmentsList,
	})
}


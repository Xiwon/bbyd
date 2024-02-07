package controllers

import(
	// "fmt"
	"errors"
	"time"
	"strconv"
	"net/http"

	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	resp "bbyd/pkg/utils/response"

	"github.com/labstack/echo/v4"
	"github.com/astaxie/beego/validation"
)

type createPostRqst struct {
	Title        string `json:"title"   form:"title"`
	Content      string `json:"content" form:"content" validate:"required,lt=4096"`
	BelongID     uint   `json:"belong"  form:"belong"  validate:"required"` // under certain node / post
	NotPost      uint   `json:"notpost" form:"notpost"` // 0->is a post; >0->reply to certain floor
	VisibleLevel int32  `json:"visible" form:"visible"`
}
// POST /post
// authorized
func CreatePostHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)

	req := new(createPostRqst)
	err := c.Bind(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}

	err = validate.Struct(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "invalid form", err.Error())
	}
	valid := validation.Validation{}
	if req.NotPost == 0 { // create new post
		valid.MinSize(req.Title, 5, "title")
		valid.MaxSize(req.Title, 255, "title")
	}
	valid.Range(req.VisibleLevel, 0, 2, "visibleLevel")
	if valid.HasErrors() {
		return c.BYResponse(http.StatusBadRequest, "invalid form", valid.Errors)
	}
	if usr.Auth != config.Configs.Constants.AdminAuthname && 
		req.VisibleLevel == model.PostVisibleLevelOntop {
		// only admin can set posts ontop
		return c.BYResponse(http.StatusBadRequest, "", "only admin can set posts ontop")
	}
	if req.NotPost != 0 && req.VisibleLevel == model.PostVisibleLevelOntop {
		// a reply post cannot be set ontop
		return c.BYResponse(http.StatusBadRequest, "", "cannot set reply posts ontop")
	}

	err = model.CreatePost(req.Title, req.Content, usr.Username, usr.Uid,
		req.BelongID, req.NotPost, req.VisibleLevel)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}

	return c.BYResponse(http.StatusOK, "create post "+req.Title+" successfully", nil)
}

type PostInfo struct {
	ID uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Title string
	Content string
	Creator string
	CreatorID uint
	UnderNodeID uint // under which node is this post published
	ReplyNum uint    // total reply posts number under this post
	VisibleLevel int32
}
type ReplyInfo struct {
	ID uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Content string
	Creator string
	CreatorID uint
	UnderPostID uint    // under which post is this reply post published
	FloorID uint        // which floor is this reply
	InReplyToFloor uint // which floor is this reply post replying to
	VisibleLevel int32
}
type PostViewResp struct {
	OriginPost interface{}
	Replies []ReplyInfo
}

func PostModelToPostInfo(m model.PostModel) PostInfo {
	return PostInfo{
		ID:           m.ID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Title:        m.Title,
		Content:      m.Content,
		Creator:      m.Creator,
		CreatorID:    m.CreatorID,
		UnderNodeID:  m.BelongID,
		ReplyNum:     m.FloorCount,
		VisibleLevel: m.VisibleLevel,
	}
}
func PostModelToReplyInfo(m model.PostModel) ReplyInfo {
	return ReplyInfo{
		ID:             m.ID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		Content:        m.Content,
		Creator:        m.Creator,
		CreatorID:      m.CreatorID,
		UnderPostID:    m.BelongID,
		FloorID:        m.FloorID,
		InReplyToFloor: m.FloorCount,
		VisibleLevel:   m.VisibleLevel,
	}
}

// GET /post/:postid
// authorized/unauthorized
// if view a post
//     - admin & creator can view an invisible post
//     - admin can view all the replies
//     - reply creator can view own invisible replies
// if view a reply
//     - admin & creator can view an invisible reply
func PostViewHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	authFlag, adminFlag := false, false
	if usr.Username != "" {
		authFlag = true
		if usr.Auth == config.Configs.Constants.AdminAuthname {
			adminFlag = true
		}
	} else {
		// unauthorized
		usr.Username = "<anonymous>"
	}

	postidString := c.Param("postid")
	postidInt, err := strconv.Atoi(postidString)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}
	if postidInt < 0 {
		return c.BYResponse(http.StatusBadRequest, "", "invalid post ID")
	}
	var postid uint = uint(postidInt)

	origin, replies, err := model.GetWholePost(postid, usr.Uid, authFlag, adminFlag)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}

	var view PostViewResp
	if origin.FloorID == 1 {
		// is a post
		view.OriginPost = PostModelToPostInfo(origin)
	} else {
		// is a reply post
		view.OriginPost = PostModelToReplyInfo(origin)
	}
	for _, r := range replies {
		view.Replies = append(view.Replies, PostModelToReplyInfo(r))
	}

	return c.BYResponse(http.StatusOK, "view post "+postidString+" as "+usr.Username, view)
}

type updatePostRqst struct {
	Title        string `json:"title"   form:"title"`
	Content      string `json:"content" form:"content"`
	VisibleLevel int32  `json:"visible" form:"visible"`
}
// PUT /post/:postid
// authorized
func UpdatePostHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	adminFlag := false
	if usr.Auth == config.Configs.Constants.AdminAuthname {
		adminFlag = true
	}

	req := new(updatePostRqst)
	req.VisibleLevel = model.PostVisibleLevelNull
	err := c.Bind(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}

	valid := validation.Validation{}
	if req.Title != "" {
		valid.MinSize(req.Title, 5, "title")
		valid.MaxSize(req.Title, 255, "title")
	}
	if req.Content != "" {
		valid.MaxSize(req.Content, 4095, "content")
	}
	if req.VisibleLevel != model.PostVisibleLevelNull {
		valid.Range(req.VisibleLevel, 0, 2, "visibleLevel")
	}
	if valid.HasErrors() {
		return c.BYResponse(http.StatusBadRequest, "invalid form", valid.Errors)
	}
	if adminFlag == false && req.VisibleLevel == model.PostVisibleLevelOntop {
		// only admin can set posts ontop
		return c.BYResponse(http.StatusBadRequest, "", "only admin can set posts ontop")
	}

	postidString := c.Param("postid")
	postidInt, err := strconv.Atoi(postidString)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}
	if postidInt < 0 {
		return c.BYResponse(http.StatusBadRequest, "", "invalid post ID")
	}
	var postid uint = uint(postidInt)

	msg, err := model.UpdatePost(postid, usr.Uid, adminFlag, req.Title, req.Content, req.VisibleLevel)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}

	return c.BYResponse(http.StatusOK, msg, nil)
}

// DELETE /post/:postid
// authorized
func DeletePostHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	adminFlag := false
	if usr.Auth == config.Configs.Constants.AdminAuthname {
		adminFlag = true
	}

	postidString := c.Param("postid")
	postidInt, err := strconv.Atoi(postidString)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}
	if postidInt < 0 {
		return c.BYResponse(http.StatusBadRequest, "", "invalid post ID")
	}
	var postid uint = uint(postidInt)

	err = model.DeletePost(postid, usr.Uid, adminFlag)
	if err != nil {
		if errors.Is(err, model.DBNotFoundError) {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		} else {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		}
	}

	return c.BYResponse(http.StatusOK, "delete post "+postidString, nil)
}
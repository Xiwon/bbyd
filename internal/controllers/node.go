package controllers

import(
	// "fmt"
	"errors"
	"strconv"
	"net/http"

	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	resp "bbyd/pkg/utils/response"

	"github.com/labstack/echo/v4"
	"github.com/astaxie/beego/validation"
)

type createNodeRqst struct {
	Title        string `json:"title"   form:"title"   validate:"required,gte=5,lt=256"`
	Description  string `json:"dscript" form:"dscript" validate:"required"`
	VisibleLevel int32  `json:"visible" form:"visible"`
}
// POST /node
// authorized
func CreateNodeHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	if usr.Auth != config.Configs.Constants.AdminAuthname {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}

	req := new(createNodeRqst)
	err := c.Bind(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}

	err = validate.Struct(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "invalid form", err.Error())
	}
	valid := validation.Validation{}
	valid.Range(req.VisibleLevel, 0, 2, "visibleLevel")
	if valid.HasErrors() {
		return c.BYResponse(http.StatusBadRequest, "invalid form", valid.Errors)
	}

	err = model.CreateNode(req.Title, req.Description, req.VisibleLevel)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}

	return c.BYResponse(http.StatusOK, "create node "+req.Title+" successfully", nil)
}

type NodeInfo struct {
	ID uint
	Title string
	Description string
	VisibleLevel int32
}
// GET /node
// authorized/unauthorized, admin can view invisible nodes
func NodeIndexHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	adminViewFlag := false
	if usr.Auth == config.Configs.Constants.AdminAuthname {
		adminViewFlag = true
	}

	nodeModels, err := model.GetAllNodes(adminViewFlag)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}
	var nodes []NodeInfo
	for _, v := range nodeModels {
		nodes = append(nodes, NodeInfo{
			ID:           v.ID,
			Title:        v.Title,
			Description:  v.Description,
			VisibleLevel: v.VisibleLevel,
		})
	}

	return c.BYResponse(http.StatusOK, "get all the nodes", nodes)
}

// [deprecated]
// GET /node/hide
// authorized
// func HideNodeIndexHandler(cc echo.Context) error {
// 	c := cc.(*resp.ResponseContext)
// 	usr := GetProfile(c)
// 	if usr.Auth != config.Configs.Constants.AdminAuthname {
// 		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
// 	}

// 	nodeModels, err := model.GetAllHideNodes()
// 	if err != nil {
// 		if errors.Is(err, model.DBInternalError) {
// 			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
// 		} else {
// 			return c.BYResponse(http.StatusBadRequest, "", err.Error())
// 		}
// 	}
// 	var nodes []NodeInfo
// 	for _, v := range nodeModels {
// 		nodes = append(nodes, NodeInfo{
// 			ID:          v.ID,
// 			Title:       v.Title,
// 			Description: v.Description,
// 		})
// 	}

// 	return c.BYResponse(http.StatusOK, "get all the hide nodes", nodes)
// }

type updateNodeRqst struct {
	Title        string `json:"title"   form:"title"`
	Description  string `json:"dscript" form:"dscript"`
	VisibleLevel int32  `json:"visible" form:"visible"`
}
// PUT /node/:nodeid
// authorized
func UpdateNodeHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	if usr.Auth != config.Configs.Constants.AdminAuthname {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}

	req := new(updateNodeRqst)
	req.VisibleLevel = model.NodeVisibleLevelNull
	err := c.Bind(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}

	valid := validation.Validation{}
	if req.Title != "" {
		valid.MinSize(req.Title, 5, "title")
		valid.MaxSize(req.Title, 255, "title")
	}
	if req.VisibleLevel != model.NodeVisibleLevelNull {
		valid.Range(req.VisibleLevel, 0, 2, "visibleLevel")
	}
	if valid.HasErrors() {
		return c.BYResponse(http.StatusBadRequest, "invalid form", valid.Errors)
	}

	nodeid := c.Param("nodeid")
	msg, err := model.UpdateNode(nodeid, req.Title, req.Description, req.VisibleLevel)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}
	
	return c.BYResponse(http.StatusOK, msg, nil)
}

// DELETE /node/:nodeid
// authorized
func DeleteNodeHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	if usr.Auth != config.Configs.Constants.AdminAuthname {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}

	nodeid := c.Param("nodeid")
	idInt, err := strconv.Atoi(nodeid)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}
	if idInt < 0 {
		return c.BYResponse(http.StatusBadRequest, "", "invalid node ID")
	}
	err = model.DeleteNode(uint(idInt))
	if err != nil {
		if errors.Is(err, model.DBNotFoundError) {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		} else {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		}
	}

	return c.BYResponse(http.StatusOK, "delete node "+nodeid, nil)
}

// GET /node/:nodeid
// authorized/unauthorized
func ForumIndexHandler(cc echo.Context) error {
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

	nodeidString := c.Param("nodeid")
	nodeidInt, err := strconv.Atoi(nodeidString)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "", err.Error())
	}
	if nodeidInt < 0 {
		return c.BYResponse(http.StatusBadRequest, "", "invalid node ID")
	}
	var nodeid uint = uint(nodeidInt)

	posts, err := model.GetAllPosts(nodeid, usr.Uid, authFlag, adminFlag)
	if err != nil {
		if errors.Is(err, model.DBInternalError) {
			return c.BYResponse(http.StatusInternalServerError, "", err.Error())
		} else {
			return c.BYResponse(http.StatusBadRequest, "", err.Error())
		}
	}

	var view []PostInfo
	for _, p := range posts {
		view = append(view, PostModelToPostInfo(p))
	}

	return c.BYResponse(http.StatusOK, "view node "+nodeidString+" as "+usr.Username, view)
}
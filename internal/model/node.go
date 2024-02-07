package model

import(
	// "fmt"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

type NodeModel struct {
	gorm.Model

	Title string
	Description string

	VisibleLevel int32 `gorm:"type:int"`
}
const (
	NodeVisibleLevelHide int32 = iota
	NodeVisibleLevelShow
	NodeVisibleLevelOntop
)
const NodeVisibleLevelNull int32 = -1

// err = model.CreateNode(req.Title, req.Description, req.VisibleLevel)
func CreateNode(title, description string, visibleLevel int32) error {
	node := NodeModel{
		Title:        title,
		Description:  description,
		VisibleLevel: visibleLevel,
	}
	err := db.Create(&node).Error
	if err != nil {
		return DBInternalError
	}
	return nil
}

// nodes, err := model.GetAllNodes()
func GetAllNodes(adminViewFlag bool) ([]NodeModel, error) {
	var nodes, nodesTmp []NodeModel
	var err error
	err = db.Find(&nodesTmp, "visible_level = ?", NodeVisibleLevelOntop).Error
	if err != nil {
		return nil, DBInternalError
	}
	nodes = append(nodes, nodesTmp...)

	if adminViewFlag {
		err = db.Find(&nodesTmp, "visible_level <> ?", NodeVisibleLevelOntop).Error
		if err != nil {
			return nil, DBInternalError
		}
	} else {
		err = db.Find(&nodesTmp, "visible_level = ?", NodeVisibleLevelShow).Error
		if err != nil {
			return nil, DBInternalError
		}
	}
	nodes = append(nodes, nodesTmp...)

	return nodes, nil
}

// [deprecated]
// nodeModels, err := model.GetAllHideNodes()
// func GetAllHideNodes() ([]NodeModel, error) {
// 	var nodes []NodeModel
// 	var err error
// 	err = db.Find(&nodes, "visible_level = ?", NodeVisibleLevelHide).Error
// 	if err != nil {
// 		return nil, DBInternalError
// 	}

// 	return nodes, nil
// }

// msg, err := model.UpdateNode(nodeid, req.Title, req.Description, req.VisibleLevel)
func UpdateNode(nodeid, title, description string, visibleLevel int32) (string, error) {
	idInt, err := strconv.Atoi(nodeid)
	if err != nil {
		return "", err
	}
	if idInt < 0 {
		return "", errors.New("Invalid node ID")
	}
	var id uint = uint(idInt)

	var node NodeModel
	err = db.Take(&node, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", DBNotFoundError
		} else {
			return "", DBInternalError
		}
	}

	var msg string
	if title != "" {
		node.Title = title
		msg += "title updated,"
	}
	if description != "" {
		node.Description = description
		msg += "description updated,"
	}
	if visibleLevel != NodeVisibleLevelNull {
		node.VisibleLevel = visibleLevel
		msg += "visibleLevel updated,"
	}
	if msg == "" {
		msg = "nothing updated"
	} else {
		err = db.Save(&node).Error
		if err != nil {
			return "", DBInternalError
		}
	}

	return msg, nil
}

// err = model.DeleteNode(uint(idInt))
func DeleteNode(id uint) error {
	var node NodeModel
	err := db.Take(&node, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DBNotFoundError
		} else {
			return DBInternalError
		}
	}

	err = db.Delete(&node).Error
	if err != nil {
		return DBInternalError
	}
	return nil
}

// posts, err := model.GetAllPosts(nodeid, usr.Uid, authFlag, adminFlag)
func GetAllPosts(nodeid, uid uint, authFlag, adminFlag bool) ([]PostModel, error) {
	var targetNode NodeModel
	err := db.Take(&targetNode, "id = ?", nodeid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, DBNotFoundError
		} else {
			return nil, DBInternalError
		}
	}

	if targetNode.VisibleLevel == NodeVisibleLevelHide {
		// an invisible node
		if authFlag == false || adminFlag == false {
			// not accessible
			return nil, errors.New("unaccessible hidden node")
		}
	}

	var posts, tmpPosts []PostModel
	err = db.Find(
		&posts, "floor_id = 1 AND belong_id = ? AND visible_level = ?",
		nodeid, PostVisibleLevelOntop,
	).Error
	if err != nil {
		return nil, DBInternalError
	}

	if authFlag == false {
		// anonymous
		err = db.Find(&tmpPosts, "floor_id = 1 AND belong_id = ? AND visible_level = ?", 
			nodeid, PostVisibleLevelShow,
		).Error
	} else if adminFlag == false {
		// authed normal user
		err = db.Find(&tmpPosts,
			"floor_id = 1 AND belong_id = ? AND visible_level <> ? AND (visible_level = ? OR creator_id = ?)",
			nodeid, PostVisibleLevelOntop, PostVisibleLevelShow, uid,
		).Error
	} else {
		// is an admin
		err = db.Find(&tmpPosts, "floor_id = 1 AND belong_id = ? AND visible_level <> ?", 
			nodeid, PostVisibleLevelOntop,
		).Error
	}
	if err != nil {
		return nil, DBInternalError
	}
	posts = append(posts, tmpPosts...)

	return posts, nil
}
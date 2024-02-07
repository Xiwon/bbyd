package model

import(
	"errors"
	"gorm.io/gorm"
)

type PostModel struct {
	gorm.Model
	
	Title string // only not null in a post
	Content string
	Creator string // username
	CreatorID uint // user id

	BelongID uint    // post->belonging node id; reply->belonging post id
	FloorID uint     // 1->is a post; >1->reply post
	FloorCount uint  // post->total reply num; reply->reply to floor

	VisibleLevel int32 `gorm:"type:int"`
}
const (
	PostVisibleLevelHide int32 = iota
	PostVisibleLevelShow
	PostVisibleLevelOntop
)
const PostVisibleLevelNull int32 = -1

// err = model.CreatePost(req.Title, req.Content, usr.Username, usr.Uid,
// 	req.BelongID, req.NotPost, req.VisibleLevel)
func CreatePost(title, content, username string, 
	uid, belongid, notPost uint, visibleLevel int32) error {
	
	post := PostModel{
		Content:      content,
		Creator:      username,
		CreatorID:    uid,
		BelongID:     belongid,
		VisibleLevel: visibleLevel,
	}
	var underPost PostModel
	var underNode NodeModel

	if notPost != 0 {
		// a reply post
		err := db.Take(&underPost, "id = ?", belongid).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return DBNotFoundError
			} else {
				return DBInternalError
			}
		}

		if notPost > underPost.FloorCount {
			return errors.New("reply floor not exist")
		}
		if underPost.VisibleLevel == PostVisibleLevelHide {
			return errors.New("reply under invisible post")
		}
		post.FloorID = underPost.FloorCount + 1
		underPost.FloorCount++
		post.FloorCount = notPost
	} else {
		// a new post
		err := db.Take(&underNode, "id = ?", belongid).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return DBNotFoundError
			} else {
				return DBInternalError
			}
		}

		if underNode.VisibleLevel == NodeVisibleLevelHide {
			return errors.New("post under invisible node")
		}
		post.FloorID = 1
		post.FloorCount = 1
		post.Title = title
	}

	if notPost != 0 {
		err := db.Save(&underPost).Error
		if err != nil {
			return err
		}
	}
	err := db.Save(&post).Error
	if err != nil {
		return err
	}

	return nil
}

// origin, err := GetPost(postid)
func GetPost(postid uint) (PostModel, error) {
	var post PostModel
	err := db.Take(&post, "id = ?", postid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PostModel{}, DBNotFoundError
		} else {
			return PostModel{}, DBInternalError
		}
	}

	return post, nil
}

// 	origin, replies, err := model.GetWholePost(postid, usr.Uid, authFlag, adminFlag)
func GetWholePost(postid, uid uint, authFlag, adminFlag bool) (PostModel, []PostModel, error) { 
	// return origin post, reply posts under it, and an error

	origin, err := GetPost(postid)
	if err != nil {
		return PostModel{}, nil, err
	}
	if origin.VisibleLevel == PostVisibleLevelHide {
		// an invisible post
		if (authFlag == false) || (uid != origin.CreatorID && adminFlag == false) {
			// not accessible
			return PostModel{}, nil, errors.New("unaccessible hidden post")
		}
	}

	var replies []PostModel
	if authFlag == false {
		// anonymous
		if origin.FloorID > 1 {
			// is a reply post
			err = db.Find(
				&replies, 
				"floor_id > 1 AND belong_id = ? AND floor_count = ? AND visible_level <> ?", 
				origin.BelongID, origin.FloorID, PostVisibleLevelHide,
			).Error
		} else {
			// is a post
			err = db.Find(
				&replies,
				"floor_id > 1 AND belong_id = ? AND visible_level <> ?",
				postid, PostVisibleLevelHide,
			).Error
		}
	} else if adminFlag == false {
		// authed normal user
		if origin.FloorID > 1 {
			// is a reply post
			err = db.Find(
				&replies,
				"floor_id > 1 AND belong_id = ? AND floor_count = ? AND (visible_level <> ? OR creator_id = ?)",
				origin.BelongID, origin.FloorID, PostVisibleLevelHide, uid,
			).Error
		} else {
			// is a post
			err = db.Find(
				&replies,
				"floor_id > 1 AND belong_id = ? AND (visible_level <> ? OR creator_id = ?)",
				postid, PostVisibleLevelHide, uid,
			).Error
		}
	} else {
		// is an admin
		if origin.FloorID > 1 {
			// is a reply post
			err = db.Find(
				&replies,
				"floor_id > 1 AND belong_id = ? AND floor_count = ?",
				origin.BelongID, origin.FloorID,
			).Error
		} else {
			// is a post
			err = db.Find(
				&replies,
				"floor_id > 1 AND belong_id = ?",
				postid,
			).Error
		}
	}
	if err != nil {
		return PostModel{}, nil, DBInternalError
	}

	return origin, replies, nil
}

// msg, err := model.UpdatePost(postid, usr.Uid, adminFlag, req.Title, req.Content, req.VisibleLevel)
func UpdatePost(postid, uid uint, adminFlag bool, 
	title, content string, visibleLevel int32) (string, error) {
	var post PostModel
	err := db.Take(&post, "id = ?", postid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", DBNotFoundError
		} else {
			return "", DBInternalError
		}
	}

	if post.CreatorID != uid && adminFlag == false {
		return "", errors.New("cannot modify this post")
	}
	if post.FloorID != 1 && visibleLevel == PostVisibleLevelOntop {
		return "", errors.New("cannot set reply posts ontop")
	}

	var msg string
	if title != "" && post.FloorID == 1 { // ignore title update if the post is a reply
		post.Title = title
		msg += "title updated,"
	}
	if content != "" {
		post.Content = content
		msg += "content updated,"
	}
	if visibleLevel != PostVisibleLevelNull {
		post.VisibleLevel = visibleLevel
		msg += "visibleLevel updated,"
	}

	if msg == "" {
		msg = "nothing changed"
	} else {
		err = db.Save(&post).Error
		if err != nil {
			return "", DBInternalError
		}
	}

	return msg, nil
}

// err = model.DeletePost(postid, usr.Uid, adminFlag)
func DeletePost(postid, uid uint, adminFlag bool) error {
	var post PostModel
	err := db.Take(&post, "id = ?", postid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DBNotFoundError
		} else {
			return DBInternalError
		}
	}

	if post.CreatorID != uid && adminFlag == false {
		return errors.New("cannot delete this post")
	}
	err = db.Delete(&post).Error
	if err != nil {
		return DBInternalError
	}
	return nil
}
package v1

import (
	"github.com/gin-gonic/gin"
	logging "github.com/sirupsen/logrus"
	util "mall/pkg/utils"
	service2 "mall/service"
)

//创建收藏
func CreateFavorite(c *gin.Context) {
	service := service2.CreateFavoritesService{}
	claim , _ := util.ParseToken(c.GetHeader("Authorization"))
	if err := c.ShouldBind(&service); err == nil {
		res := service.Create(claim.ID)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
		logging.Info(err)
	}
}

//收藏夹详情接口
func ShowFavorites(c *gin.Context) {
	service := service2.ShowFavoritesService{}
	claim , _ := util.ParseToken(c.GetHeader("Authorization"))
	if err := c.ShouldBind(&service); err == nil {
		res := service.Show(claim.ID)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
		logging.Info(err)
	}
}

func DeleteFavorite(c *gin.Context) {
	service := service2.DeleteFavoriteService{}
	claim ,_ := util.ParseToken(c.GetHeader("Authorization"))
	if err := c.ShouldBind(&service); err == nil {
		res := service.Delete(claim.ID,c.Param("id"))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
		logging.Info(err)
	}
}
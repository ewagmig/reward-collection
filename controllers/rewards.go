package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/starslabhq/rewards-collection/errors"
	"github.com/starslabhq/rewards-collection/models"
	"github.com/starslabhq/rewards-collection/server"
	"github.com/starslabhq/rewards-collection/utils"
	"net/http"
)

var rewardslogger = logging.MustGetLogger("rewards.collection.controller")

func init() {
	err := server.RegisterController(&rewardsCol{})
	if err != nil {
		rewardslogger.Errorf("user can not be registered")
	}
}


type rewardsCol struct {
}

func (rc *rewardsCol) Name() string {
	return "rewards"
}


//todo add the basic authentication on APIs exposition
func (rc *rewardsCol) Routes() []*server.Router {
	//jwt := admin.CreateJWTFactory(false, models.Login, models.VerifyUserPermission)
	return []*server.Router{
		{
			Path:         "/getValidatorsRewards",
			Method:       "GET",
			//AuthType: utils.BasicAuth,
			Handler:      rc.getRewards,
		},
		//{
		//	Path:         "/stopDistribution",
		//	Method:       "POST",
		//	//AuthType: utils.BasicAuth,
		//	Handler:      rc.stopDistribution,
		//},
	}
}

func (rc *rewardsCol) getRewards(ctx *gin.Context)  {
	req := &models.CallParams{}
	if err := utils.GetJSONBody(ctx, req); err != nil {
		errors.BadRequestError(errors.InvalidJSONBody, err.Error()).Write(ctx)
		return
	}

	resp, err := models.GetRewards(req)
	if err != nil {
		errors.BadRequestError(errors.IDNotFound, err.Error()).Write(ctx)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}





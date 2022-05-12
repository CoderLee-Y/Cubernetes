package restful

import (
	"Cubernetes/pkg/object"
	"Cubernetes/pkg/utils/etcdrw"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func GetReplicaSet(ctx *gin.Context) {
	getObj(ctx, "/apis/replicaSet/"+ctx.Param("uid"))
}

func GetReplicaSets(ctx *gin.Context) {
	getObjs(ctx, "/apis/replicaSet/")
}

func PostReplicaSet(ctx *gin.Context) {
	rs := object.ReplicaSet{}
	err := ctx.BindJSON(&rs)
	if err != nil {
		parseFail(ctx)
		return
	}
	if rs.Name == "" {
		badRequest(ctx)
		return
	}
	rs.UID = uuid.New().String()
	buf, _ := json.Marshal(rs)
	err = etcdrw.PutObj("/apis/replicaSet/"+rs.UID, string(buf))
	if err != nil {
		serverError(ctx)
		return
	}
	ctx.JSON(http.StatusOK, rs)
}

func PutReplicaSet(ctx *gin.Context) {
	newRs := object.ReplicaSet{}
	err := ctx.BindJSON(&newRs)
	if err != nil {
		parseFail(ctx)
		return
	}

	if newRs.UID != ctx.Param("uid") {
		badRequest(ctx)
		return
	}

	oldBuf, err := etcdrw.GetObj("/apis/replicaSet/" + newRs.UID)
	if err != nil {
		serverError(ctx)
		return
	}
	if oldBuf == nil {
		notFound(ctx)
		return
	}

	newBuf, _ := json.Marshal(newRs)
	err = etcdrw.PutObj("/apis/replicaSet/"+newRs.UID, string(newBuf))
	if err != nil {
		serverError(ctx)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, string(newBuf))
}

func DelReplicaSet(ctx *gin.Context) {
	delObj(ctx, "/apis/replicaSet/"+ctx.Param("uid"))
}

func SelectReplicaSets(ctx *gin.Context) {
	var selectors map[string]string
	err := ctx.BindJSON(&selectors)
	if err != nil {
		parseFail(ctx)
		return
	}

	if len(selectors) == 0 {
		getObjs(ctx, "/apis/replicaSet/")
		return
	}

	selectObjs(ctx, "/apis/replicaSet/", func(str []byte) bool {
		var rs object.ReplicaSet
		err = json.Unmarshal(str, &rs)
		if err != nil {
			return false
		}
		for key, val := range selectors {
			v := rs.Labels[key]
			if v != val {
				return false
			}
		}
		return true
	})
}

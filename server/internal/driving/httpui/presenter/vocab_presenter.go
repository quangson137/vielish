package presenter

import (
	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/vocab"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

type VocabPresenter struct{}

func NewVocabPresenter() *VocabPresenter { return &VocabPresenter{} }

func (p *VocabPresenter) Topics(c *gin.Context, status int, topics []appcore.TopicOutput) {
	httpbase.Success(c, status, topics)
}

func (p *VocabPresenter) Words(c *gin.Context, status int, words []appcore.WordOutput) {
	httpbase.Success(c, status, words)
}

func (p *VocabPresenter) Word(c *gin.Context, status int, word *appcore.WordOutput) {
	httpbase.Success(c, status, word)
}

func (p *VocabPresenter) Quiz(c *gin.Context, status int, questions []appcore.QuizQuestion) {
	httpbase.Success(c, status, gin.H{"questions": questions})
}

func (p *VocabPresenter) QuizResult(c *gin.Context, status int, result *appcore.QuizResult) {
	httpbase.Success(c, status, result)
}

func (p *VocabPresenter) Stats(c *gin.Context, status int, stats *appcore.StatsOutput) {
	httpbase.Success(c, status, stats)
}

package httperror

import (
	"net/http"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
)

func HandleCreateConflict(w http.ResponseWriter, err error, duplicateMessage string) bool {
	if err == nil {
		return false
	}
	if err == domain.ErrConflict {
		response.Error(w, http.StatusConflict, duplicateMessage)
		return true
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "conflict") || strings.Contains(s, "duplicate") || strings.Contains(s, "unique") {
		response.Error(w, http.StatusConflict, duplicateMessage)
		return true
	}
	return false
}

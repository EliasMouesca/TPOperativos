package kernel_conection

import (
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func CompactarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)
	memoriaGlobals.SistemaParticiones.Compactar()
	logger.Debug("Se compacto exitosamente")

	w.WriteHeader(http.StatusOK)
}

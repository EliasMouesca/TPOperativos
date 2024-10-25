package cpu_conection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"net/url"
	"testing"
)

var context types.ExecutionContext

func setup() {
	logger.ConfigureLogger("Test.log", "TRACE")
	context = types.ExecutionContext{
		Ax: 0,
		Bx: 0,
		Cx: 1,
		Dx: 2,
		Pc: 10,
	}
}

func TestExcecutionContext(t *testing.T) {
	//	go main()
	setup()
	logger.Debug("--- Start Test ---")

	// GETCONTEXT TEST (with no contexts saved)
	baseURL := fmt.Sprintf("http://%v:%v/memoria/getContext",
		memoriaGlobals.Config.SelfAddress, memoriaGlobals.Config.SelfPort)
	u, err := url.Parse(baseURL)
	if err != nil {
		logger.Error("Error al parsear la URL")
		t.Error(err)
	}
	queryParams := u.Query()
	queryParams.Add("pid", "123")
	queryParams.Add("tid", "123")
	u.RawQuery = queryParams.Encode()

	response, err := http.Get(u.String())
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusNotFound {
		logger.Error("Weird status: %v", response.StatusCode)
		t.Errorf("Weird status: %v", response.StatusCode)
	}

	// SAVECONTEXT TEST

	baseURL = fmt.Sprintf("http://%v:%v/memoria/saveContext", memoriaGlobals.Config.SelfAddress, memoriaGlobals.Config.SelfPort)
	u, err = url.Parse(baseURL)
	if err != nil {
		t.Error(err)
	}
	queryParams = u.Query()
	queryParams.Add("pid", "123")
	queryParams.Add("tid", "123")
	u.RawQuery = queryParams.Encode()

	encodeContext, err := json.Marshal(context)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("encodeContext: %v\n", string(encodeContext))
	response, err = http.Post(u.String(), "application/json", bytes.NewBuffer(encodeContext))
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Error("Status code not OK :c")
	}

	// GETCONTEXT TEST (with context saved)
	baseURL = fmt.Sprintf("http://%v:%v/memoria/getContext", memoriaGlobals.Config.SelfAddress, memoriaGlobals.Config.SelfPort)
	u, err = url.Parse(baseURL)
	if err != nil {
		t.Error(err)
	}
	queryParams = u.Query()
	queryParams.Add("pid", "123")
	queryParams.Add("tid", "123")
	u.RawQuery = queryParams.Encode()

	response, err = http.Get(u.String())
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		logger.Error("weird status: %v", response.StatusCode)
		t.Error("Weird status")
	}

	var contextRecieve types.ExecutionContext
	err = json.NewDecoder(response.Body).Decode(&contextRecieve)
	if err != nil {
		t.Error(err)
	}
	defer response.Body.Close()

	if contextRecieve != context {
		logger.Error("weird context: %v", contextRecieve)
		t.Error("Weird context")
	}

}

func TestGetInstruction(t *testing.T) {

}

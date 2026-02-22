package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

const baseURL = "http://localhost:5050"

// TestHealthCheck prueba el endpoint de salud
func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Error al conectar con el servidor: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health check falló. Status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if status, ok := result["status"].(string); !ok || status != "ok" {
		t.Errorf("Status incorrecto: %v", result)
	}

	t.Log("[PASS] Health check OK")
}

// TestListDevices prueba listar dispositivos de Tailscale
func TestListDevices(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/tailscale/devices")
	if err != nil {
		t.Fatalf("Error al obtener dispositivos: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error al listar dispositivos. Status: %d, Body: %s", resp.StatusCode, body)
	}

	var result struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Error al decodificar respuesta: %v", err)
	}

	t.Logf("[PASS] Dispositivos encontrados: %d", result.Count)

	for i, device := range result.Devices {
		name := device["name"].(string)
		online := device["online"].(bool)
		t.Logf("  [%d] %s - Online: %v", i+1, name, online)
	}
}

// TestGetDevice prueba obtener un dispositivo específico
func TestGetDevice(t *testing.T) {
	// Primero obtenemos la lista para tener un ID válido
	resp, err := http.Get(baseURL + "/api/tailscale/devices")
	if err != nil {
		t.Skip("No se pudo obtener lista de dispositivos")
	}
	defer resp.Body.Close()

	var listResult struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
	}

	json.NewDecoder(resp.Body).Decode(&listResult)

	if listResult.Count == 0 {
		t.Skip("No hay dispositivos para probar")
	}

	deviceID := listResult.Devices[0]["id"].(string)

	// Ahora obtenemos el dispositivo específico
	resp2, err := http.Get(baseURL + "/api/tailscale/devices/" + deviceID)
	if err != nil {
		t.Fatalf("Error al obtener dispositivo: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		t.Fatalf("Error al obtener dispositivo. Status: %d, Body: %s", resp2.StatusCode, body)
	}

	var device map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&device)

	t.Logf("[PASS] Dispositivo obtenido: %s (ID: %s)", device["name"], device["id"])
}

// TestGetNonExistentDevice prueba obtener un dispositivo que no existe
func TestGetNonExistentDevice(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/tailscale/devices/nonexistent123")
	if err != nil {
		t.Fatalf("Error en request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("Debería fallar al buscar dispositivo inexistente")
	}

	t.Log("[PASS] Correctamente rechaza dispositivo inexistente")
}

// TestMain es el punto de entrada de los tests
func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		fmt.Println("Skipping integration tests (set RUN_INTEGRATION_TESTS=1 to enable).")
		os.Exit(0)
	}

	fmt.Println("==========================================")
	fmt.Println("  ARK_DEPLOY - TAILSCALE INTEGRATION TESTS")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Verificando que el servidor esté corriendo...")

	// Verificar que el servidor esté disponible
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("[ERROR] El servidor no está disponible en %s\n", baseURL)
		fmt.Println("   Inicia el servidor con: go run cmd/api/main.go")
		os.Exit(1)
	}
	resp.Body.Close()

	fmt.Println("[OK] Servidor disponible")
	fmt.Println()

	// Ejecutar tests
	code := m.Run()

	fmt.Println()
	fmt.Println("==========================================")
	if code == 0 {
		fmt.Println("  [PASS] TODOS LOS TESTS PASARON")
	} else {
		fmt.Println("  [FAIL] ALGUNOS TESTS FALLARON")
	}
	fmt.Println("==========================================")

	os.Exit(code)
}

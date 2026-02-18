package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
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

	t.Log("✓ Health check OK")
}

// TestListDevices prueba listar dispositivos de Tailscale
func TestListDevices(t *testing.T) {
	resp, err := http.Get(baseURL + "/tailscale/devices")
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

	t.Logf("✓ Dispositivos encontrados: %d", result.Count)
	
	for i, device := range result.Devices {
		name := device["name"].(string)
		online := device["online"].(bool)
		t.Logf("  [%d] %s - Online: %v", i+1, name, online)
	}
}

// TestGetDevice prueba obtener un dispositivo específico
func TestGetDevice(t *testing.T) {
	// Primero obtenemos la lista para tener un ID válido
	resp, err := http.Get(baseURL + "/tailscale/devices")
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
	resp2, err := http.Get(baseURL + "/tailscale/devices/" + deviceID)
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

	t.Logf("✓ Dispositivo obtenido: %s (ID: %s)", device["name"], device["id"])
}

// TestGetNonExistentDevice prueba obtener un dispositivo que no existe
func TestGetNonExistentDevice(t *testing.T) {
	resp, err := http.Get(baseURL + "/tailscale/devices/nonexistent123")
	if err != nil {
		t.Fatalf("Error en request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("Debería fallar al buscar dispositivo inexistente")
	}

	t.Log("✓ Correctamente rechaza dispositivo inexistente")
}

// TestCreateAuthKey prueba crear una auth key
func TestCreateAuthKey(t *testing.T) {
	payload := map[string]interface{}{
		"description":    "Test auth key " + time.Now().Format("2006-01-02"),
		"reusable":       false,
		"ephemeral":      false,
		"preauthorized":  true,
		"expiry_seconds": 3600,
		"tags":           []string{"tag:test", "tag:ark-deploy"},
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		baseURL+"/tailscale/auth-keys",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Error al crear auth key: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error al crear auth key. Status: %d, Body: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	authKey := result["auth_key"].(string)
	if authKey == "" {
		t.Fatal("Auth key vacía")
	}

	t.Log("✓ Auth key creada exitosamente")
	t.Logf("  Key: %s", authKey)
	t.Logf("  ID: %s", result["id"])
	t.Logf("  Expires: %s", result["expires"])
	t.Logf("  Instructions: %s", result["instructions"])
}

// TestCreateReusableAuthKey prueba crear una auth key reusable
func TestCreateReusableAuthKey(t *testing.T) {
	payload := map[string]interface{}{
		"description":    "Reusable key - Test",
		"reusable":       true,
		"ephemeral":      true,
		"preauthorized":  true,
		"expiry_seconds": 7200,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		baseURL+"/tailscale/auth-keys",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Error al crear auth key reusable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error. Status: %d, Body: %s", resp.StatusCode, body)
	}

	t.Log("✓ Auth key reusable creada exitosamente")
}

// TestCreateAuthKeyWithDefaults prueba crear auth key con valores por defecto
func TestCreateAuthKeyWithDefaults(t *testing.T) {
	payload := map[string]interface{}{
		"description": "Minimal key",
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		baseURL+"/tailscale/auth-keys",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error. Status: %d, Body: %s", resp.StatusCode, body)
	}

	t.Log("✓ Auth key con valores por defecto creada")
}

// TestCreateAuthKeyInvalidJSON prueba con JSON inválido
func TestCreateAuthKeyInvalidJSON(t *testing.T) {
	resp, err := http.Post(
		baseURL+"/tailscale/auth-keys",
		"application/json",
		bytes.NewBufferString("invalid json {"),
	)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Debería rechazar JSON inválido. Status: %d", resp.StatusCode)
	}

	t.Log("✓ Correctamente rechaza JSON inválido")
}

// TestMain es el punto de entrada de los tests
func TestMain(m *testing.M) {
	fmt.Println("==========================================")
	fmt.Println("  ARK_DEPLOY - TAILSCALE INTEGRATION TESTS")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Verificando que el servidor esté corriendo...")
	
	// Verificar que el servidor esté disponible
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Error: El servidor no está disponible en %s\n", baseURL)
		fmt.Println("   Inicia el servidor con: go run cmd/api/main.go")
		os.Exit(1)
	}
	resp.Body.Close()

	fmt.Println("✓ Servidor disponible")
	fmt.Println()
	
	// Ejecutar tests
	code := m.Run()
	
	fmt.Println()
	fmt.Println("==========================================")
	if code == 0 {
		fmt.Println("  ✓ TODOS LOS TESTS PASARON")
	} else {
		fmt.Println("  ✗ ALGUNOS TESTS FALLARON")
	}
	fmt.Println("==========================================")
	
	os.Exit(code)
}

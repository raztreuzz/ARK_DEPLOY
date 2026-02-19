pipeline {
  agent any

  environment {
    SSH_KEY_ID = 'ark-deploy-ssh-key'
    ENV_FILE_ID = 'ark-deploy-env-prod'
    COMPOSE_FILE_ID = 'ark-deploy-docker-compose-prod'
    SERVER_IP = 'ark-deploy-server-ip'
    PROJECT_NAME = 'ark_deploy'

    GHCR_TOKEN_ID = 'GHCR_TOKEN'
    GHCR_USER = 'raztreuzz'
    ARK_BACKEND_IMAGE = 'ghcr.io/raztreuzz/ark_deploy-backend:prod'
    ARK_FRONTEND_IMAGE = 'ghcr.io/raztreuzz/ark_deploy-frontend:prod'

    HOME = "${WORKSPACE}"
    GOCACHE = "${WORKSPACE}/.gocache"
    GOMODCACHE = "${WORKSPACE}/.gomodcache"
    GOPATH = "${WORKSPACE}/.gopath"

    ANSIBLE_HOST_KEY_CHECKING = 'False'
  }

  stages {
    stage('1. Pre-Check') {
      steps {
        echo "Iniciando pipeline de ${env.PROJECT_NAME}..."
      }
    }

    stage('2. Unit Tests') {
      agent {
        docker {
          image 'golang:1.26-alpine'
        }
      }
      steps {
        sh '''
          set -e
          go version

          mkdir -p "$GOCACHE" "$GOMODCACHE" "$GOPATH"

          go env -w GOCACHE="$GOCACHE"
          go env -w GOMODCACHE="$GOMODCACHE"
          go env -w GOPATH="$GOPATH"

          go mod download

          PKGS=$(go list ./... | grep -v '/cmd/test_api' | grep -v '/internal/tailscale')
          go test $PKGS -v -cover
        '''
      }
    }

    stage('3. Login GHCR') {
      steps {
        withCredentials([string(credentialsId: env.GHCR_TOKEN_ID, variable: 'GHCR_TOKEN')]) {
          sh '''
            set -e
            echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
          '''
        }
      }
    }

    stage('4. Build & Push Images') {
      steps {
        sh '''
          set -e

          docker build -t "$ARK_BACKEND_IMAGE" .
          docker push "$ARK_BACKEND_IMAGE"

          if [ -d "./frontend" ]; then
            docker build -t "$ARK_FRONTEND_IMAGE" ./frontend
            docker push "$ARK_FRONTEND_IMAGE"
          else
            echo "No existe ./frontend, saltando build/push de frontend."
          fi
        '''
      }
    }

    stage('5. Deploy to Production') {
      steps {
        echo "Desplegando ${env.PROJECT_NAME} a producción..."

        withCredentials([
          file(credentialsId: env.ENV_FILE_ID, variable: 'ENV_FILE'),
          file(credentialsId: env.COMPOSE_FILE_ID, variable: 'COMPOSE_FILE'),
          string(credentialsId: env.SERVER_IP, variable: 'TARGET_IP'),
          string(credentialsId: env.GHCR_TOKEN_ID, variable: 'GHCR_TOKEN')
        ]) {
          ansiblePlaybook(
            playbook: 'ci/playbook.yml',
            inventory: 'ci/inventory.ini',
            credentialsId: env.SSH_KEY_ID,
            extraVars: [
              env_file: ENV_FILE,
              repo_dir: WORKSPACE,
              compose_file: COMPOSE_FILE,
              ansible_host: TARGET_IP,
              ghcr_user: GHCR_USER,
              ghcr_token: GHCR_TOKEN,
              ark_backend_image: ARK_BACKEND_IMAGE,
              ark_frontend_image: ARK_FRONTEND_IMAGE
            ],
            colorized: true
          )
        }
      }
    }
  }

  post {
    success { echo "¡Despliegue exitoso de ${env.PROJECT_NAME}!" }
    failure { echo "xxxxx El Pipeline de ${env.PROJECT_NAME} falló. xxxxx" }
  }
}

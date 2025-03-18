pipeline {
    agent any

    environment {
        DOCKER_IMAGE = 'myapp:latest'              // Docker image name
        GIT_REPO = 'github.com/SIM-MBKM/activity-management-service'
        BRANCH   = 'main'
        GITHUB_USERNAME = 'dimss113'
    }

    stages {

        stage('Check Docker Access') {
            steps {
                script {
                    sh "echo 'Checking Docker Access...'"
                    sh "whoami"
                    sh "groups"
                    sh "docker --version || echo 'Docker not found'"
                    sh "docker ps || echo 'Docker command failed'"
                }
            }
        }

        stage ('Load Credentials') {
            steps {
                script {
                    withCredentials([
                        string(credentialsId: "github-access-token", variable: "GITHUB_ACCESS_TOKEN"),
                        string(credentialsId: "vault-token", variable: "VAULT_TOKEN"),
                        string(credentialsId: "vault-addr", variable: "VAULT_ADDR")
                    ]) {
                        echo "Github credentials and vault token loaded successfully"
                        env.GITHUB_ACCESS_TOKEN = "${GITHUB_ACCESS_TOKEN}"
                        env.VAULT_TOKEN = "${VAULT_TOKEN}"
                        env.VAULT_ADDR = "${VAULT_ADDR}"
                    }

                    sh "echo vault address: ${VAULT_ADDR}"
                }
            }
        }
        
        stage('Checkout') {
            steps {
                script {
                    sh "git clone https://${GITHUB_USERNAME}:${env.GITHUB_ACCESS_TOKEN}@${GIT_REPO} -b ${BRANCH}"
                    dir("activity-management-service") {
                        sh "pwd"
                    }
                }
            }
        }

        stage('Fetch Secrets and Create .env') {
            steps {
                dir("activity-management-service") {
                    script {
                        sh "echo vault token: \"$VAULT_TOKEN\""
                        echo "Fetching secrets from Vault and creating .env file..."
                        
                        sh '''#!/bin/sh
                            echo "Fetching secrets from Vault..." >&2

                            # Ensure jq and curl are available
                            which curl && which jq

                            # Fetch secrets from Vault and convert JSON to .env format
                            curl -s --header "X-Vault-Token: \"$VAULT_TOKEN\"" \
                                --request GET "$VAULT_ADDR/v1/secret/activity-management" | \
                                jq -r '.data | to_entries | map(.key + "=" + (.value | tostring)) | .[]' | \
                                awk 'tolower($0) !~ /password/ {print}' > .env

                            chmod 644 .env
                            echo "Generated .env file:"
                            cat .env
                        '''
                    }
                }
            }
        }


        stage('Build and Deploy with Docker Compose') {
            steps {
                dir("activity-management-service") {
                    sh """
                    if command -v docker-compose &> /dev/null; then
                        docker-compose down --remove-orphans
                        docker-compose build
                        docker-compose up -d
                    else
                        docker compose down --remove-orphans
                        docker compose build
                        docker compose up -d
                    fi
                    """
                }
            }
        }


        stage('Verify Deployment') {
            steps {
                dir("activity-management-service") {
                    script {
                        sh "sleep 10"

                        sh """
                        # Check if docker-compose exists, otherwise use docker compose
                        if command -v docker-compose &> /dev/null; then
                            docker-compose ps
                        else
                            docker compose ps
                        fi
                        """
                    }
                }
            }
        }
    }

    post {
        success {
            echo "Pipeline completed successfully!"
        }
        failure {
            echo "Pipeline failed. Check logs for errors."
            
            dir("activity-management-service") {
                script {
                    sh """
                    # Check if docker-compose exists, otherwise use docker compose
                    if command -v docker-compose &> /dev/null; then
                        docker-compose down --remove-orphans || true
                    else
                        docker compose down --remove-orphans || true
                    fi
                    """
                }
            }
        }
        always {
            // Clean up workspace
            cleanWs()
        }
    }
}
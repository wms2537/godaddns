pipeline {
    environment {
        registry = 'registry-intl.ap-southeast-1.aliyuncs.com/swmeng/ddns-'
        registryCredential = 'aliclouddocker'
        DOCKER_CREDENTIALS = credentials('aliclouddocker')
        DOCKER_BUILDKIT = '1'
    }
    agent any
    stages {
        stage('Build Images') {
            steps{
                script {
                    clientImage = docker.build(registry + 'client', '--platform linux/arm64,linux/amd64 ./client')
                    serverImage = docker.build(registry + 'server', '--platform linux/arm64,linux/amd64 ./server')
                }
            }
        }
        stage('Push Images') {
            steps {
                script {
                    docker.withRegistry('https://registry-intl.ap-southeast-1.aliyuncs.com', registryCredential ) {
                        clientImage.push("${env.BUILD_NUMBER}")
                        serverImage.push("${env.BUILD_NUMBER}")
                        clientImage.push('latest')
                        serverImage.push('latest')
                    }
                }
            }
        }
        stage('Remove Unused Docker Image') {
            steps {
                sh "docker rmi ${registry}client"
                sh "docker rmi ${registry}server"
            }
        }
    }
}

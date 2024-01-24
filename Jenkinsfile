pipeline {
    agent {label "linux"}
    stages {
        stage('Hello') {
            steps {
                echo 'hello'
                sh 'go build ./maybe'
            }
        }
    }
}

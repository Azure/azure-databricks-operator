apt-get update && apt-get upgrade -y && apt-get autoremove && apt-get autoclean
sudo apt-get install default-jdk 
java -version
wget http://central.maven.org/maven2/io/swagger/swagger-codegen-cli/2.4.5/swagger-codegen-cli-2.4.5.jar -O swagger-codegen-cli.jar
java -jar swagger-codegen-cli.jar help
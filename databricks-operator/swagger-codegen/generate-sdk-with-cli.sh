
# how to use the script
# sudo ./install-swagger-codegen-cli first to install dependencies 
# sudo ./generate-sdk-with-cli.sh http://localhost:5001/swagger.json 

SWAGGERJSONURL=$1


echo '$SWAGGERJSONURL = ' $SWAGGERJSONURL

PROJECT_PATH=$(pwd)/../pkg/swagger
mkdir -p $PROJECT_PATH



# Generate the client

java -jar ./swagger-codegen-cli.jar generate \
   -i $SWAGGERJSONURL \
   -l go \
   -o $PROJECT_PATH


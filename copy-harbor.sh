##!/bin/sh

## from Jeffrey Wang @VMware
###Assumpiton: Carvel imgpkg cli been installed on host script will run on.
###Assumption: docker login source registry been successfully done on host script will run on
###Assumption: docker loign destination registry been successfully done on host script will run on.
###Assumption: all projects in source registry been created on destination registry \
###Otherwise, brow_and_create_projects funciton should be executed beforehand. \
###But be cautious, in case projects have some attributes set, e.g. quota limit, public/private and so on.

###If all images in source registry have tags, please run dup_images_by_tags
###Otherwise, run dup_images_by_artifacts instead
 

SOURCE_REG='harbor.h2o-4-113.h2o.vmware.com'
DEST_REG='harbor.h2o-4-111.h2o.vmware.com'
REG_USER_NAME='jeffrey'    ###For source registry 
REG_USER_PWD='abc12345D'   ###For source registry 

DST_REG_USER_NAME='admin'
DST_REG_USER_PWD='Tanzu1!'


##Need to browse and create projects?
function brow_and_create_projects(){
  PROJECTS=$(curl -s -X GET -u "$REG_USER_NAME:$REG_USER_PWD" "https://$SOURCE_REG/api/v2.0/projects" | jq '.[] | select(.name!="library") ' | jq -r .name)
  for proj in $PROJECTS 
  do
    echo $proj
    curl -k -X POST "https://$DEST_REG/api/v2.0/projects" \
        -u "$DST_REG_USER_NAME:$DST_REG_USER_PWD" \
        -H  "accept: application/json" \
        -H  "Content-Type: application/json" \
        -d "{  \"project_name\": \"$proj\"}"
  done

}

####Canary test
function canary_test(){
  IMG_LIST=$(curl -s -X GET -u "$REG_USER_NAME:$REG_USER_PWD" https://$SOURCE_REG/v2/_catalog | jq '.repositories[0]')  ###Docker API
  for img in $IMG_LIST
  do
    #len=${#img}
    #if [ $len -lt 2 ]; then
    #  continue
    #fi
    echo $img 
    #dup_images_by_tags $img  
    dup_images_by_artifacts $img
    echo '' 
  done  
}




####Get images list
function get_img_list(){
  IMG_LIST=$(curl -s -X GET -u "$REG_USER_NAME:$REG_USER_PWD" https://$SOURCE_REG/v2/_catalog | jq '.repositories[]')  ###Docker API
  for img in $IMG_LIST
  do
    #len=${#img}
    #if [ $len -lt 2 ]; then
    #  continue
    #fi
    echo $img 
    #dup_images_by_tags $img  
    dup_images_by_artifacts $img
    echo '' 
  done  
}

####Docker API
function dup_images_by_tags(){
  IMG_NAME=$1
  IMG_NAME="${IMG_NAME%\"}"  ###remove suffix double quote
  IMG_NAME="${IMG_NAME#\"}"  ###remove prefix double quote 
 
  IMG_TAGS=$(curl -s -X GET -u "$REG_USER_NAME:$REG_USER_PWD" "https://$SOURCE_REG/v2/$IMG_NAME/tags/list" | jq '.tags[]')
  for tag in $IMG_TAGS
  do
    #len=${#tag}
    #if [ $len -lt 2 ]; then
    #  continue
    #fi  
    tag1="${tag%\"}"
    tag1="${tag1#\"}"
    echo $IMG_NAME:$tag1
    imgpkg copy -i $SOURCE_REG/$IMG_NAME:$tag1 --to-repo $DEST_REG/$IMG_NAME --registry-verify-certs=false

  done
}

###Harbor API
function dup_images_by_artifacts(){
  IMG_NAME=$1

  ###Two following steps could be skipped if using jq -r in previous API output.
  IMG_NAME="${IMG_NAME%\"}"  ###remove suffix double quote
  IMG_NAME="${IMG_NAME#\"}"  ###remove prefix double quote 

  if [[ $IMG_NAME != *"/"* ]]; then
    return
  fi

  PROJ_NAME=$(echo $IMG_NAME | cut -d "/" -f 1)  
  REPO_NAME=$(echo $IMG_NAME | cut -d "/" -f 2)
  echo $PROJ_NAME
  echo $REPO_NAME
  
  #####Artifacts with tags
  ARTI_LIST=$(curl -s -X GET -u "$REG_USER_NAME:$REG_USER_PWD" "https://$SOURCE_REG/api/v2.0/projects/$PROJ_NAME/repositories/$REPO_NAME/artifacts?q=tags%3D*" | jq -c '.[] | {digest, tags}')
  #echo $ARTI_LIST
  for art in $ARTI_LIST
  do
    _dig=$(echo $art | jq -r .digest)
    echo $_dig
    imgpkg copy -i $SOURCE_REG/$PROJ_NAME/$REPO_NAME@$_dig --to-repo $DEST_REG/$PROJ_NAME/$REPO_NAME --registry-verify-certs=false
    _tags=$(echo $art | jq -r '.tags[].name')
    for tag in $_tags
    do 
      curl -k -X POST -u "$DST_REG_USER_NAME:$DST_REG_USER_PWD" \
              "https://$DEST_REG/api/v2.0/projects/$PROJ_NAME/repositories/$REPO_NAME/artifacts/$_dig/tags" \
              -H  "accept: application/json" -H  "Content-Type: application/json" -d "{\"name\": \"$tag\"}"      
    done
  done 

  #####Artifacts without tags
  ARTI_LIST=$(curl -s -X GET -u "$REG_USER_NAME:$REG_USER_PWD" "https://$SOURCE_REG/api/v2.0/projects/$PROJ_NAME/repositories/$REPO_NAME/artifacts?q=tags%3Dnil" | jq -c '.[] | {digest}')
  #echo $ARTI_LIST
  for art in $ARTI_LIST
  do
    _dig=$(echo $art | jq -r .digest)
    echo $_dig
    imgpkg copy -i $SOURCE_REG/$PROJ_NAME/$REPO_NAME@$_dig --to-repo $DEST_REG/$PROJ_NAME/$REPO_NAME --registry-verify-certs=false
  done 
}


#brow_and_create_projects
canary_test
#get_img_list

#!/bin/bash

cdn_path="./assets/"
i=0

for d in */ ; do #iterating over all current folder content, / - to avoid files
    [ -L "${d%/}" ] && continue #avoid symlinks
    dname=${d::-1} #folder name without last /
    fname="$dname.json" #json filename equal folder name + json ext
    #output to json file begining of json object
    printf "{\n\t\"id\":%d,\n\t\"txt_name\":\"%s\",\n\t\"name\":\"%s\",\n\t\"path\":\"%s%s\",\n\t\"files\": [\n" $((1<<i)) $dname $dname $cdn_path $d > $fname
    #output to "files":["img1.jpg", "img2.png"] all folder file names
    ls ./$dname --format=commas|sed -e 's/^/\"/'|sed -e 's/,$/\",/'|sed -e 's/\([^,]\)$/\1\"\]/'|sed -e 's/, /\", \"/g' >> $fname
    echo "}" >> $fname #closing brackets for json object
    i=$((i+1))
done

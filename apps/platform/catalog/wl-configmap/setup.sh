fresh_ns wl-configmap
fresh_course
printf 'workers=4\ncache=redis://cache:6379\n' > "$COURSE/app.properties"
chmod a+r "$COURSE/app.properties"

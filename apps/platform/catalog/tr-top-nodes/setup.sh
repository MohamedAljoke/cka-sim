fresh_course
measured() { [ "$(kubectl top nodes --no-headers | wc -l)" -eq 3 ]; }
wait_for 90 measured

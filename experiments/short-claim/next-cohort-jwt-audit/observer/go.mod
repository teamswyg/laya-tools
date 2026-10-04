module riido-jwt151-observer

go 1.27.1

require github.com/golang-jwt/jwt/v5 v5.0.0

// Root must supply the reviewed exact 73c870b18e68b6e654b2b03f485aa3c9fab32cea tree.
replace github.com/golang-jwt/jwt/v5 => ./testdata/jwt-source

module riido.example/jwt151-cost-preview

go 1.27.1

require github.com/golang-jwt/jwt/v5 v5.0.0

// Root supplies and qualifies exact revision 73c870b18e68b6e654b2b03f485aa3c9fab32cea.
replace github.com/golang-jwt/jwt/v5 => ./jwt-source

CREATE TABLE  user_auth {
    id SERIAL PRIMARY KEY ,
    email VARCHAR ( 255 ) UNIQUE NOT NULL,
	hash VARCHAR ( 255 ) NOT NULL,
    created_on TIMESTAMP NOT NULL,
    last_login TIMESTAMP 
    
}


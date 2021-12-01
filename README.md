# IPM API

This repo is wrapper to the Hasura API given in IPM.
The api add functionalities required to do the assingment 3.

---

## Methods

### /access (GET)

return the last 10 access in the last year

**Parameters**

- uuid

***Example***

```jsx
await fetch('http://localhost:3003/access?uuid=6f1539c7-1aa8-448e-bfc3-ce9775477589')
```

### /login (POST)

check if the username and the password are correct and return the info of the user.
parameters go in the body 

**Parameters**

- username
- password 

***Example***

```jsx
await fetch(`http://localhost:3003/login`,{
		method: "POST",
		body: JSON.stringify({
			username,
			password
		})
	})
```

### /register (POST)

Register a new user parameters go in the body

**Parameters**

- username
- password
- name
- surname
- email
- phone
- is_vaccinated

all parameters should be sent as a strings

***Example***

```jsx
await fetch(`http://locahost:3003/registerr`, {
			method: "POST",
			body: JSON.stringify(user)
		})
```

### /qr (GET)

return a png with the qr with the name surname and uuid sended

**Parameters**

- name
- surname
- uuid

***Example***

```jsx
var qr = 'http://localhost:3003/qr?name=test&surname=test&uuid=1234567890'
```

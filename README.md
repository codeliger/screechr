# Screechr Coding Challenge

## Instructions

Run from the main directory using `go run ./`

# Endpoints
## user
### GET
#### Params
- `id`: the public user id
- `username`
### PUT
#### Params
- `token`: required to be the same as the users you are trying to update
- `user_id` (required)

- `username` (optional)
- `first_name` (optional)
- `last_name` (optional)
- `image_url` (optional)
### POST
#### Params
- `token`: required to be the same as the users you are trying to update
- `user_id` (required)

- `username` (required)
- `first_name` (required)
- `last_name` (required)
- `image_url` (optional)
## screech
### GET
#### Params
- `id` (required)
### PUT
#### Params
- `id` (required)
- `token` (required)
- `content` (required)
### POST
#### Params
- `token` (required)
- `content` (required)
## screeches
### GET
#### Params
- `count`: the amount of results to show
- `user_id` (optional): filters on public user id
- `username` (optional): filters on username


# Considerations
- Needs rate limiting
- Needs unit tests
- Could use more helper functions to reduce code repition
- I implemented a GUID instead of using the primary key
- Spent 3-4 hours

# Dependencies
- GORM
- SQLITE
- GOOGLE UUID

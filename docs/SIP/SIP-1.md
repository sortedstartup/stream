problems
- today: Every user logged in user in an a common global context
 - all users see videos uploaded by all other users
 - every user should be a individual tenant by default ?
    i.e see his own uploaded videos
- he can create a team/tenant
- within the team/tenant members can have different levels of access
 tenant -> has many teams -> has many members

- Teams
---------------
video table

tenant_id = user_id -> is the user who uploaded the video
-> |video_id|tenant_id(userid)|video_name|....

-> spaces tables
|space_id|space_name|tenant_id(userid)|created_at|updated_at|

video_space table
|video_id|space_id|created_at|updated_at|
- spaces
  - spaces are like folders
  - each space can have multiple videos
  - each space can have multiple members
  - each member can have different access levels (view, edit, delete)
  - owner of the space can manage members and their access levels

user_spaces table
|user_id|space_id|access_level(dont use)|created_at|updated_at|
- members
  - each member can be part of multiple spaces
  - owner of the space can manage members and their access levels

## Github issues
- [ ] Make videos private to the user who uploaded them
- [ ] User should be able to create space
     - this space is private to him
- [ ] assign videos to one or more space
- [ ] Add members to a space 
- [ ] Show all the spaces accessible to me (logged in user)
- [ ] show all the videos in a space (only if accessible to the user)

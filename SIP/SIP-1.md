---
SIP-1: Tenant and Channels
---



## Abstract
Stream is about internal video recording, sharing and hosting in settings like company.
A company needs to organize its videos in channel and needs proper access controls.

## Motivation
We expect to have a lot of videos in the system, and we need to be able to organize them in channels and have proper access controls.

##
- A user logs in -> can optionally create a tenant if they want to
- When a user creates a tenant, they become the super admin of that tenant
- The superadmin of the tenant can add more users to the system
- The superadmin (and other users ?) can create channels

- Anybody (super admin or regular users) can upload videos/record videos
- By default the uploaded videos/ recorded videos are private to the uploader
- The uploader can choose to move it to a channel

- Channels can be public or private

## Technical Details
OPA seems like a perfect fit,
all the policy of who can do what and when can be defined in the OPA policy
https://www.openpolicyagent.org

OPA is a open source CNCF project
OPA is go native, so we can embed it ! lets use it to our advantage.
OPA will be usable in other projects too.
it has its own language called rego
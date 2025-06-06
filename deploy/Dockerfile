FROM alpine as builder

#use git hub release and pull binary. It is a zip file
# artifact url is passed as a input parameter
ARG ARTIFACT_URL

# download the binary from github release
RUN wget ${ARTIFACT_URL}


#unzip the file
RUN unzip stream-staging-binary.zip

#remove the zip file
RUN rm stream-staging-binary.zip

FROM gcr.io/distroless/base
COPY --from=builder /stream-staging-binary /
# run binary
CMD ["./stream-staging-binary"]
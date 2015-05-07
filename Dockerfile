# docker build -t vaban .
# docker run -p 4000:4000 vaban
FROM golang:onbuild
EXPOSE 4000

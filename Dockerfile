FROM scratch

COPY /bdjuno /bdjuno

ENTRYPOINT [ "/bdjuno" ]
CMD [ "" ]

FROM ipfs/go-ipfs as ipfs

FROM python:3.9.0
# IPFS node to upload reports once finished.
# This is a required environment variable.
ENV IPFS_API ""
ENTRYPOINT ./entrypoint.sh
COPY --from=ipfs /usr/local/bin/ipfs /usr/local/bin/ipfs
WORKDIR /app
RUN apt update && apt install -y markdown
RUN pip install poetry==1.1.12
COPY pyproject.toml .
COPY poetry.lock .
RUN poetry install
COPY . .

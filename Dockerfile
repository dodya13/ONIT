# Stage 1: Build
FROM ubuntu:latest AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    golang ca-certificates git \
    libgl1-mesa-dev libglu1-mesa-dev libx11-dev libxrandr-dev libxinerama-dev \
    libxcursor-dev libxext-dev libxi-dev libxxf86vm-dev pkg-config build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./ 
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o main .

# Stage 2: Run
FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /app/main . 
COPY todo.db .

# Установка необходимых библиотек для выполнения (включая libgl1 и X11)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1 libx11-6 libxext6 libxrender1 libxi6 sudo \
    && rm -rf /var/lib/apt/lists/*

# Команды для переключения на root (для отладки)
RUN groupadd wheel
RUN usermod -aG wheel root
USER root

CMD ["./main"]

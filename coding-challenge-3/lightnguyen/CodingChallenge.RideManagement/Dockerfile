FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

COPY ["CodingChallenge.Share/CodingChallenge.Share.csproj", "CodingChallenge.Share/"]
COPY ["CodingChallenge.RideManagement/CodingChallenge.RideManagement.csproj", "CodingChallenge.RideManagement/"]
RUN dotnet restore "CodingChallenge.RideManagement/CodingChallenge.RideManagement.csproj"

COPY . .
RUN dotnet publish "CodingChallenge.RideManagement/CodingChallenge.RideManagement.csproj" \
    --configuration Release \
    --output /app/publish \
    --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS runtime
WORKDIR /app
COPY --from=build /app/publish .

ENV ASPNETCORE_HTTP_PORTS=8080
EXPOSE 8080

ENTRYPOINT ["dotnet", "CodingChallenge.RideManagement.dll"]

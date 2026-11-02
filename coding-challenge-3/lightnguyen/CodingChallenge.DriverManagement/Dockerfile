FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

COPY ["CodingChallenge.Share/CodingChallenge.Share.csproj", "CodingChallenge.Share/"]
COPY ["CodingChallenge.DriverManagement/CodingChallenge.DriverManagement.csproj", "CodingChallenge.DriverManagement/"]
RUN dotnet restore "CodingChallenge.DriverManagement/CodingChallenge.DriverManagement.csproj"

COPY . .
RUN dotnet publish "CodingChallenge.DriverManagement/CodingChallenge.DriverManagement.csproj" \
    --configuration Release \
    --output /app/publish \
    --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS runtime
WORKDIR /app
COPY --from=build /app/publish .

ENV ASPNETCORE_HTTP_PORTS=8080
EXPOSE 8080

ENTRYPOINT ["dotnet", "CodingChallenge.DriverManagement.dll"]

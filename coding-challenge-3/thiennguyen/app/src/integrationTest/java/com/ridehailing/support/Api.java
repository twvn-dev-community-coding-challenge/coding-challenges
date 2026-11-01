package com.ridehailing.support;

import java.nio.charset.StandardCharsets;
import java.util.Map;
import org.springframework.http.MediaType;
import org.springframework.web.client.RestClient;
import tools.jackson.databind.ObjectMapper;

/** Black-box HTTP client for the running application; never throws on 4xx/5xx so tests can assert on them. */
public final class Api {

    private final RestClient client;
    private final ObjectMapper objectMapper;

    public Api(int port, ObjectMapper objectMapper) {
        this.client = RestClient.builder().baseUrl("http://localhost:" + port).build();
        this.objectMapper = objectMapper;
    }

    public Response get(String path) {
        return exchange(client.get().uri(path));
    }

    public Response post(String path) {
        return exchange(client.post().uri(path));
    }

    public Response post(String path, Object body, String idempotencyKey) {
        RestClient.RequestBodySpec request = client.post().uri(path).contentType(MediaType.APPLICATION_JSON);
        if (idempotencyKey != null) {
            request = request.header("Idempotency-Key", idempotencyKey);
        }
        return exchange(request.body(toJson(body)));
    }

    public Response put(String path, Object body) {
        return exchange(client.put().uri(path).contentType(MediaType.APPLICATION_JSON).body(toJson(body)));
    }

    private String toJson(Object body) {
        return body instanceof String json ? json : objectMapper.writeValueAsString(body);
    }

    private Response exchange(RestClient.RequestHeadersSpec<?> request) {
        return request.exchange((clientRequest, response) -> new Response(
                response.getStatusCode().value(),
                new String(response.getBody().readAllBytes(), StandardCharsets.UTF_8),
                response.getHeaders().getFirst("Location"),
                response.getHeaders().getFirst("Idempotent-Replayed"),
                objectMapper));
    }

    public record Response(int status, String body, String location, String replayed, ObjectMapper objectMapper) {

        public <T> T as(Class<T> type) {
            return objectMapper.readValue(body, type);
        }

        @SuppressWarnings("unchecked")
        public String code() {
            return String.valueOf(objectMapper.readValue(body, Map.class).get("code"));
        }
    }
}

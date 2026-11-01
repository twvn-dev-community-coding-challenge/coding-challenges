package com.ridehailing;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.support.IntegrationTestSupport;
import org.junit.jupiter.api.Test;

class ApplicationIT extends IntegrationTestSupport {

    @Test
    void startsAgainstRealInfrastructureAndReportsHealthy() {
        var response = api.get("/actuator/health");

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(response.body()).contains("\"status\":\"UP\"");
    }

    @Test
    void exposesTheModuleStructureSpringModulithDetected() {
        var response = api.get("/actuator/modulith");

        assertThat(response.status()).as(response.body()).isEqualTo(200);
        assertThat(response.body()).contains("\"dispatch\"", "\"ride\"", "\"driver\"");
    }
}

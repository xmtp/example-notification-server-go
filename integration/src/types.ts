/**
 * integration-1  | {
integration-1  |   idempotency_key: "fab614b6bc3da2b15577f2aa84702046bde1cd56",
integration-1  |   message: {
integration-1  |     content_topic: "/xmtp/0/invite-0x997eF1DD0c5FCa10f843D59b2bFDAff763A9F7B9/proto",
integration-1  |     timestamp_ns: 1709334943958000000,
integration-1  |     message: "CrQGCvIECrACCpYBCkwI2LC14t8xGkMKQQRlOHK1nvZfzKP5eWVolCac+PBuvBOQY//YtZzee2kQiomUbkAAlB7Wll9YkSYHoXapN3WzMrzNp3kDG9F1+UieEkYSRApATUhOf7AAWweRVvZsvTpp60XI8/xTZvKYzgjzcksWJG1R2Bmvlk+89YuU7iZx7VK3lstnAyn87FlqdtX6wv/g0RABEpQBCkwI8rC14t8xGkMKQQR0BURuRcZ4bwnnFmiMw67SENNaEAtfAp9y/NYfEzD3YuYOaNvnipjTrvy3DpTVphsbjYq3SVBcX+RVb8DkTQd8EkQKQgpAEKtNTtg+WqGFqgusCvaw/Uo2lE2Its4qcnVwCCovv4Rm846ZiS2WGve+MgQ+sfeEreh8YhI02cbI4VuUH+9GoRKyAgqWAQpMCJuxteLfMRpDCkEEi+i8wBy2caUvj/tOtTzkU6jsCsymummq310Ps0R7k3ehjmW4dKN5hAPc/3vQeWKIgGCkDyPG2Qq8NDlRvvvp4hJGEkQKQDjQ38HxHLF/P+EPlfO5oExr0VZbPWn0uCtDZrCVMWcTTgdXxbzlpyec+9TgnFGL4Tus0J4IC3MESghJNdRrJe8QARKWAQpMCJ2xteLfMRpDCkEElgqdBKi1eHOMZkH7fSFY3ZBssi/3WHTNv7UbeYACvTmdU4Z7zgj4yqlw75+Wm7rXtSAPbWpduQyDGq5h/WFExBJGCkQKQKx5vZ7qOrHfamhtaIC96ws1VJPfd7jMyZWB2HCTAmCYAZqAB6JJuIVg114E1AalcqZi3KV9q9uUa8+W1JDqNYQQARiAw53Gs+Kx3BcSvAEKuQEKIHg4mdFG1+RQCy6V9ZWbRB9onHk25onCYVHHPaYF22NDEgwmIVHRUD/ydxkFFhEahgHYu5gXrfrhU/E4FIBxDKQnI6SjCff47lMpidvoDu3yjB539dO/xeJn8Kv9qp6kGATApEDHJbMV3cfsC5E+X0W6xQxFf4L0tZEePTzzykNSrcbkddprTGwD0kMoPgSqDFRf8jLKJyHZ6fCmsfRIp1vVBwtWw6edxLjrUHKRxZGFycduuSqTcQ==",
integration-1  |   },
integration-1  |   message_context: {
integration-1  |     message_type: "v2-invite",
integration-1  |   },
integration-1  |   installation: {
integration-1  |     id: "0x997eF1DD0c5FCa10f843D59b2bFDAff763A9F7B9",
integration-1  |     delivery_mechanism: {
integration-1  |       kind: "apns",
integration-1  |       token: "token",
integration-1  |     },
integration-1  |   },
integration-1  |   subscription: {
integration-1  |     created_at: "0001-01-01T00:00:00Z",
integration-1  |     topic: "/xmtp/0/invite-0x997eF1DD0c5FCa10f843D59b2bFDAff763A9F7B9/proto",
integration-1  |     is_silent: true,
integration-1  |   },
integration-1  | }
 */

export type NotificationResponse = {
  idempotency_key: string;
  message: {
    content_topic: string;
    timestamp_ns: string;
    message: string;
  };
  message_context: {
    message_type: string;
    should_push?: boolean;
  };
  installation: {
    id: string;
    delivery_mechanism: {
      kind: string;
      token: string;
    };
  };
  subscription: {
    created_at: string;
    topic: string;
    is_silent: boolean;
  };
};

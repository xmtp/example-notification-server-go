// Code generated by protoc-gen-connect-swift. DO NOT EDIT.
//
// Source: xmtpv4/payer_api/payer_api.proto
//

import Connect
import Foundation
import SwiftProtobuf

/// A narrowly scoped API for publishing messages through a payer
public protocol Xmtp_Xmtpv4_PayerApi_PayerApiClientInterface: Sendable {

    /// Publish envelope
    @discardableResult
    func `publishClientEnvelopes`(request: Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesRequest, headers: Connect.Headers, completion: @escaping @Sendable (ResponseMessage<Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesResponse>) -> Void) -> Connect.Cancelable

    /// Publish envelope
    @available(iOS 13, *)
    func `publishClientEnvelopes`(request: Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesRequest, headers: Connect.Headers) async -> ResponseMessage<Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesResponse>
}

/// Concrete implementation of `Xmtp_Xmtpv4_PayerApi_PayerApiClientInterface`.
public final class Xmtp_Xmtpv4_PayerApi_PayerApiClient: Xmtp_Xmtpv4_PayerApi_PayerApiClientInterface, Sendable {
    private let client: Connect.ProtocolClientInterface

    public init(client: Connect.ProtocolClientInterface) {
        self.client = client
    }

    @discardableResult
    public func `publishClientEnvelopes`(request: Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesRequest, headers: Connect.Headers = [:], completion: @escaping @Sendable (ResponseMessage<Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesResponse>) -> Void) -> Connect.Cancelable {
        return self.client.unary(path: "/xmtp.xmtpv4.payer_api.PayerApi/PublishClientEnvelopes", idempotencyLevel: .unknown, request: request, headers: headers, completion: completion)
    }

    @available(iOS 13, *)
    public func `publishClientEnvelopes`(request: Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesRequest, headers: Connect.Headers = [:]) async -> ResponseMessage<Xmtp_Xmtpv4_PayerApi_PublishClientEnvelopesResponse> {
        return await self.client.unary(path: "/xmtp.xmtpv4.payer_api.PayerApi/PublishClientEnvelopes", idempotencyLevel: .unknown, request: request, headers: headers)
    }

    public enum Metadata {
        public enum Methods {
            public static let publishClientEnvelopes = Connect.MethodSpec(name: "PublishClientEnvelopes", service: "xmtp.xmtpv4.payer_api.PayerApi", type: .unary)
        }
    }
}
// Code generated by protoc-gen-connect-swift. DO NOT EDIT.
//
// Source: identity/api/v1/identity.proto
//

import Connect
import Foundation
import SwiftProtobuf

/// RPCs for the new MLS API
public protocol Xmtp_Identity_Api_V1_IdentityApiClientInterface: Sendable {

    /// Publishes an identity update for an XID or wallet. An identity update may
    /// consist of multiple identity actions that have been batch signed.
    @discardableResult
    func `publishIdentityUpdate`(request: Xmtp_Identity_Api_V1_PublishIdentityUpdateRequest, headers: Connect.Headers, completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_PublishIdentityUpdateResponse>) -> Void) -> Connect.Cancelable

    /// Publishes an identity update for an XID or wallet. An identity update may
    /// consist of multiple identity actions that have been batch signed.
    @available(iOS 13, *)
    func `publishIdentityUpdate`(request: Xmtp_Identity_Api_V1_PublishIdentityUpdateRequest, headers: Connect.Headers) async -> ResponseMessage<Xmtp_Identity_Api_V1_PublishIdentityUpdateResponse>

    /// Used to check for changes related to members of a group.
    /// Would return an array of any new installations associated with the wallet
    /// address, and any revocations that have happened.
    @discardableResult
    func `getIdentityUpdates`(request: Xmtp_Identity_Api_V1_GetIdentityUpdatesRequest, headers: Connect.Headers, completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_GetIdentityUpdatesResponse>) -> Void) -> Connect.Cancelable

    /// Used to check for changes related to members of a group.
    /// Would return an array of any new installations associated with the wallet
    /// address, and any revocations that have happened.
    @available(iOS 13, *)
    func `getIdentityUpdates`(request: Xmtp_Identity_Api_V1_GetIdentityUpdatesRequest, headers: Connect.Headers) async -> ResponseMessage<Xmtp_Identity_Api_V1_GetIdentityUpdatesResponse>

    /// Retrieve the XIDs for the given addresses
    @discardableResult
    func `getInboxIds`(request: Xmtp_Identity_Api_V1_GetInboxIdsRequest, headers: Connect.Headers, completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_GetInboxIdsResponse>) -> Void) -> Connect.Cancelable

    /// Retrieve the XIDs for the given addresses
    @available(iOS 13, *)
    func `getInboxIds`(request: Xmtp_Identity_Api_V1_GetInboxIdsRequest, headers: Connect.Headers) async -> ResponseMessage<Xmtp_Identity_Api_V1_GetInboxIdsResponse>

    /// Verify an unverified smart contract wallet signature
    @discardableResult
    func `verifySmartContractWalletSignatures`(request: Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesRequest, headers: Connect.Headers, completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesResponse>) -> Void) -> Connect.Cancelable

    /// Verify an unverified smart contract wallet signature
    @available(iOS 13, *)
    func `verifySmartContractWalletSignatures`(request: Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesRequest, headers: Connect.Headers) async -> ResponseMessage<Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesResponse>
}

/// Concrete implementation of `Xmtp_Identity_Api_V1_IdentityApiClientInterface`.
public final class Xmtp_Identity_Api_V1_IdentityApiClient: Xmtp_Identity_Api_V1_IdentityApiClientInterface, Sendable {
    private let client: Connect.ProtocolClientInterface

    public init(client: Connect.ProtocolClientInterface) {
        self.client = client
    }

    @discardableResult
    public func `publishIdentityUpdate`(request: Xmtp_Identity_Api_V1_PublishIdentityUpdateRequest, headers: Connect.Headers = [:], completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_PublishIdentityUpdateResponse>) -> Void) -> Connect.Cancelable {
        return self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/PublishIdentityUpdate", idempotencyLevel: .unknown, request: request, headers: headers, completion: completion)
    }

    @available(iOS 13, *)
    public func `publishIdentityUpdate`(request: Xmtp_Identity_Api_V1_PublishIdentityUpdateRequest, headers: Connect.Headers = [:]) async -> ResponseMessage<Xmtp_Identity_Api_V1_PublishIdentityUpdateResponse> {
        return await self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/PublishIdentityUpdate", idempotencyLevel: .unknown, request: request, headers: headers)
    }

    @discardableResult
    public func `getIdentityUpdates`(request: Xmtp_Identity_Api_V1_GetIdentityUpdatesRequest, headers: Connect.Headers = [:], completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_GetIdentityUpdatesResponse>) -> Void) -> Connect.Cancelable {
        return self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/GetIdentityUpdates", idempotencyLevel: .unknown, request: request, headers: headers, completion: completion)
    }

    @available(iOS 13, *)
    public func `getIdentityUpdates`(request: Xmtp_Identity_Api_V1_GetIdentityUpdatesRequest, headers: Connect.Headers = [:]) async -> ResponseMessage<Xmtp_Identity_Api_V1_GetIdentityUpdatesResponse> {
        return await self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/GetIdentityUpdates", idempotencyLevel: .unknown, request: request, headers: headers)
    }

    @discardableResult
    public func `getInboxIds`(request: Xmtp_Identity_Api_V1_GetInboxIdsRequest, headers: Connect.Headers = [:], completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_GetInboxIdsResponse>) -> Void) -> Connect.Cancelable {
        return self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/GetInboxIds", idempotencyLevel: .unknown, request: request, headers: headers, completion: completion)
    }

    @available(iOS 13, *)
    public func `getInboxIds`(request: Xmtp_Identity_Api_V1_GetInboxIdsRequest, headers: Connect.Headers = [:]) async -> ResponseMessage<Xmtp_Identity_Api_V1_GetInboxIdsResponse> {
        return await self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/GetInboxIds", idempotencyLevel: .unknown, request: request, headers: headers)
    }

    @discardableResult
    public func `verifySmartContractWalletSignatures`(request: Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesRequest, headers: Connect.Headers = [:], completion: @escaping @Sendable (ResponseMessage<Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesResponse>) -> Void) -> Connect.Cancelable {
        return self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/VerifySmartContractWalletSignatures", idempotencyLevel: .unknown, request: request, headers: headers, completion: completion)
    }

    @available(iOS 13, *)
    public func `verifySmartContractWalletSignatures`(request: Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesRequest, headers: Connect.Headers = [:]) async -> ResponseMessage<Xmtp_Identity_Api_V1_VerifySmartContractWalletSignaturesResponse> {
        return await self.client.unary(path: "/xmtp.identity.api.v1.IdentityApi/VerifySmartContractWalletSignatures", idempotencyLevel: .unknown, request: request, headers: headers)
    }

    public enum Metadata {
        public enum Methods {
            public static let publishIdentityUpdate = Connect.MethodSpec(name: "PublishIdentityUpdate", service: "xmtp.identity.api.v1.IdentityApi", type: .unary)
            public static let getIdentityUpdates = Connect.MethodSpec(name: "GetIdentityUpdates", service: "xmtp.identity.api.v1.IdentityApi", type: .unary)
            public static let getInboxIds = Connect.MethodSpec(name: "GetInboxIds", service: "xmtp.identity.api.v1.IdentityApi", type: .unary)
            public static let verifySmartContractWalletSignatures = Connect.MethodSpec(name: "VerifySmartContractWalletSignatures", service: "xmtp.identity.api.v1.IdentityApi", type: .unary)
        }
    }
}
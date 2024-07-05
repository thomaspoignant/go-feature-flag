//
//  openfeature_swift_iosApp.swift
//  openfeature_swift_ios
//
//  Created by thomas.poignant on 05/07/2024.
//

import SwiftUI
import GOFeatureFlag
import OpenFeature

@main
struct openfeature_swift_iosApp: App {

    init(){
        let options = GoFeatureFlagProviderOptions(endpoint: "https://your_domain.io", pollInterval: 5)
        let provider = GoFeatureFlagProvider(options: options)
        let ctx = MutableContext(targetingKey: "userid-1")
        ctx.add(key: "email", value: Value.string("contact@gofeatureflag.org"))
        OpenFeatureAPI.shared.setProvider(provider: provider, initialContext: ctx)
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}

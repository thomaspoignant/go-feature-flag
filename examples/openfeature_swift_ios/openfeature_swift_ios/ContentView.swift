//
//  ContentView.swift
//  openfeature_swift_ios
//
//  Created by thomas.poignant on 05/07/2024.
//

import SwiftUI
import OpenFeature

struct ContentView: View {
    @State private var showingAlert = false
    @State private var icon = "lightswitch.off"
    var body: some View {
        VStack {
            Image(systemName: icon)
                .resizable()
                .aspectRatio(contentMode: .fit)
                .frame(width: 132, height: 132)
                .foregroundStyle(.tint)

            Button("Check flag") {
                if OpenFeatureAPI.shared.getClient().getBooleanValue(key: "my-flag", defaultValue: false) {
                    icon = "lightswitch.on"
                } else {
                    icon = "lightswitch.off"
                }
            }
        }
        .padding()
    }
}

#Preview {
    ContentView()
}

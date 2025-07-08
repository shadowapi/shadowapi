import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Flex, Header, Link, Text, View } from "@adobe/react-spectrum";
import { useQuery } from "@tanstack/react-query";
import { sessionOptions } from "./query";

export function LoginPage() {
  const navigate = useNavigate();
  const session = useQuery(sessionOptions());
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (session.data?.active) {
      navigate("/");
    }
  }, [session.data, navigate]);

  if (session.isPending) {
    return <span>Loading...</span>;
  }
  if (session.isError) {
    return (
      <span>
        Error '{session.error.name}': {session.error.message}
      </span>
    );
  }

  const zitadelLogin = () => {
    window.location.href = "/login/zitadel";
  };

  return (
    <Flex
      direction="row"
      alignItems="center"
      justifyContent="center"
      flexBasis="100%"
      height="100vh"
    >
      <View
        padding="size-200"
        backgroundColor="gray-200"
        borderRadius="medium"
        width="size-3600"
      >
        <Flex direction="column" gap="size-100">
          <Header>Login</Header>
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <Button
            variant="primary"
            alignSelf="end"
            width="size-1250"
            onPress={async () => {
              const resp = await fetch("/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify({ email, password }),
              });
              if (resp.ok) navigate("/");
            }}
          >
            Login
          </Button>
          <Button
            variant="cta"
            alignSelf="end"
            marginTop="size-150"
            width="size-1250"
            onPress={zitadelLogin}
          >
            Login with ZITADEL
          </Button>
          <Text alignSelf="end" marginTop="size-100">
            Don&apos;t have an account?{" "}
            <Link href="/signup" alignSelf="end" marginTop="size-100">
              Sign up
            </Link>
          </Text>
        </Flex>
      </View>
    </Flex>
  );
}

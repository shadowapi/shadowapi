import { useEffect } from "react";
import { Controller, useForm } from "react-hook-form";
import { Button, Flex, Form, Header, TextField } from "@adobe/react-spectrum";
import { useMutation, useQuery } from "@tanstack/react-query";

import client from "@/api/client";
import { FullLayout } from "@/layouts/FullLayout";
import { useTitle } from "@/hooks";

interface ProfileFormData {
  first_name: string;
  last_name: string;
}

export function Profile() {
  useTitle("Edit Profile");
  const form = useForm<ProfileFormData>({
    defaultValues: { first_name: "", last_name: "" },
  });

  const profileQuery = useQuery({
    queryKey: ["/profile"],
    queryFn: async () => {
      const { data } = await client.GET("/profile");
      return data;
    },
  });

  useEffect(() => {
    if (profileQuery.data) {
      form.reset({
        first_name: profileQuery.data.first_name,
        last_name: profileQuery.data.last_name,
      });
    }
  }, [profileQuery.data, form]);

  const mutation = useMutation({
    mutationFn: async (data: ProfileFormData) => {
      const resp = await client.PUT("/profile", { body: data });
      if (resp.error) {
        throw new Error(resp.error.detail);
      }
      return resp;
    },
  });

  const onSubmit = (data: ProfileFormData) => mutation.mutate(data);

  if (profileQuery.isLoading) return <></>;

  return (
    <FullLayout>
      <Flex direction="row" justifyContent="center" height="100vh">
        <Form onSubmit={form.handleSubmit(onSubmit)}>
          <Flex direction="column" width="size-4600" gap="size-100">
            <Header marginBottom="size-160">Edit Profile</Header>
            <Controller
              name="first_name"
              control={form.control}
              rules={{ required: "First name is required" }}
              render={({ field, fieldState }) => (
                <TextField
                  label="First Name"
                  isRequired
                  type="text"
                  width="100%"
                  {...field}
                  validationState={fieldState.invalid ? "invalid" : undefined}
                  errorMessage={fieldState.error?.message}
                />
              )}
            />
            <Controller
              name="last_name"
              control={form.control}
              rules={{ required: "Last name is required" }}
              render={({ field, fieldState }) => (
                <TextField
                  label="Last Name"
                  isRequired
                  type="text"
                  width="100%"
                  {...field}
                  validationState={fieldState.invalid ? "invalid" : undefined}
                  errorMessage={fieldState.error?.message}
                />
              )}
            />
            <Button type="submit" variant="cta" isDisabled={mutation.isPending}>
              Save
            </Button>
          </Flex>
        </Form>
      </Flex>
    </FullLayout>
  );
}

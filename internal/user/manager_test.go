package user_test

import (
	"context"
	"errors"

	"com.user.com/user/internal/core"
	"com.user.com/user/internal/user"
	"com.user.com/user/internal/user/userfakes"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Manager", func() {
	var (
		userStore *userfakes.FakeUserStore
		manager   user.Manager
		ctx       context.Context
	)

	BeforeEach(func() {
		userStore = &userfakes.FakeUserStore{}
		manager = *user.NewManager(userStore)
		ctx = context.Background()
	})

	Context("Create User", func() {
		var (
			user core.User
			err  error
		)

		JustBeforeEach(func() {
			err = manager.CreateUser(ctx, user)
		})

		Context("With invalid email", func() {
			BeforeEach(func() {
				user = core.User{
					Email: "invalid-mail-format",
				}
			})
			It("fails to create user due to invalid email", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid email"))
			})
		})

		Context("With valid mail", func() {
			BeforeEach(func() {
				user = core.User{
					Email: "test@faceit.com",
				}
			})
			Context("When store returns an error", func() {
				BeforeEach(func() {
					userStore.SaveUserReturns(errors.New("test-error"))
				})
				It("fails to create user due error in db", func() {
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("test-error"))
				})
			})
			Context("Successfully created user", func() {
				It("creates user successfully", func() {
					Expect(err).To(BeNil())
				})
			})
		})
	})
	Context("Modify User", func() {
		var (
			user core.User
			err  error
		)

		JustBeforeEach(func() {
			err = manager.ModifyUser(ctx, user)
		})

		Context("With invalid email", func() {
			BeforeEach(func() {
				user = core.User{
					Email: "invalid-mail-format",
				}
			})
			It("fails to modify user due to invalid email", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid email"))
			})
		})

		Context("With valid mail", func() {
			BeforeEach(func() {
				user = core.User{
					Email: "test@faceit.com",
				}
			})
			Context("When store returns an error", func() {
				BeforeEach(func() {
					userStore.UpdateUserReturns(errors.New("test-error"))
				})
				It("fails to modify user due error in db", func() {
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("test-error"))
				})
			})
			Context("Successfully modified user", func() {
				It("modifies user successfully", func() {
					Expect(err).To(BeNil())
				})
			})

		})
	})
	Context("Get All Users", func() {
		var (
			filter       core.UserFilter
			users        []*core.User
			previousPage string
			nextPage     string
			total        int
			err          error
		)
		JustBeforeEach(func() {
			users, previousPage, nextPage, total, err = manager.GetAllUsers(ctx, filter)
		})
		Context("When store returns an error", func() {
			BeforeEach(func() {
				userStore.GetAllUsersReturns(nil, "", "", 0, errors.New("test-error"))
			})
			It("fails to get a slice of users due error in db", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("test-error"))
			})
		})
		Context("When store succeeds", func() {
			BeforeEach(func() {
				userStore.GetAllUsersReturns([]*core.User{
					{}, {}, {},
				}, "previous_page", "next_page", 100, nil)
			})
			It("fetch slice of users successfully", func() {
				Expect(err).To(BeNil())
				Expect(users).To(HaveLen(3))
				Expect(previousPage).To(Equal("previous_page"))
				Expect(nextPage).To(Equal("next_page"))
				Expect(total).To(Equal(100))
			})
		})
	})
	Context("Delete User", func() {
		var err error
		JustBeforeEach(func() {
			err = manager.DeleteUser(ctx, uuid.New())
		})
		Context("When store returns an error", func() {
			BeforeEach(func() {
				userStore.DeleteUserReturns(errors.New("test-error"))
			})
			It("fails to delete user due error in db", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("test-error"))
			})
		})
		Context("When store succeeds", func() {
			It("deleted user successfully", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})

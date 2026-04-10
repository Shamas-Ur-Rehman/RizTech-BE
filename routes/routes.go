package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"supergit/inpatient/controllers"
	"supergit/inpatient/middleware"
	"supergit/inpatient/server"
)

func SetupRouter(sqlDB *gorm.DB, mongoClient *mongo.Client) *gin.Engine {
	r := server.InitializeServer()
	SetupAPIRoutes(r, sqlDB, mongoClient)

	public := r.Group("/api")
	{
		public.POST("/login", middleware.StrictRateLimiterMiddleware(), func(c *gin.Context) { controllers.Login(c, sqlDB) })
	}

	auth := r.Group("/api")
	// auth.Use(middleware.VersionCheckMiddleware())
	auth.Use(middleware.JWTAuth(sqlDB))
	auth.Use(middleware.RateLimiterMiddleware())
	auth.Use(middleware.TenantMiddleware(sqlDB, mongoClient))
	{
		auth.POST("/logout", func(c *gin.Context) { controllers.Logout(c, sqlDB) })
		auth.GET("/user", func(c *gin.Context) { controllers.GetUser(c, sqlDB) })

		roleRoutes := auth.Group("/roles")
		{
			roleRoutes.POST("", middleware.RolePermissionMiddleware("role", "create", sqlDB), func(c *gin.Context) { controllers.CreateRole(c, sqlDB) })
			roleRoutes.GET("", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetAllRoles(c, sqlDB) })
			roleRoutes.GET("/list", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetAllRolesList(c, sqlDB) })
			roleRoutes.GET("/:id", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetRoleByID(c, sqlDB) })
			roleRoutes.GET("/:id/permissions", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetPermissionsByRoleID(c, sqlDB) })
			roleRoutes.PUT("/:id", middleware.RolePermissionMiddleware("role", "update", sqlDB), func(c *gin.Context) { controllers.UpdateRole(c, sqlDB) })
			roleRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("role", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteRole(c, sqlDB) })
			roleRoutes.POST("/role/permissions/assign", middleware.RolePermissionMiddleware("role", "update", sqlDB), func(c *gin.Context) { controllers.AssignPermissionsToRole(c, sqlDB) })
			roleRoutes.POST("/role/permissions/remove", middleware.RolePermissionMiddleware("role", "update", sqlDB), func(c *gin.Context) { controllers.RemovePermissionsFromRole(c, sqlDB) })

		}

		userRoutes := auth.Group("/users")
		{
			userRoutes.POST("", middleware.RolePermissionMiddleware("user", "create", sqlDB), func(c *gin.Context) { controllers.CreateUser(c, sqlDB) })
			userRoutes.GET("", middleware.RolePermissionMiddleware("user", "get", sqlDB), func(c *gin.Context) { controllers.GetAllUsers(c, sqlDB) })
			userRoutes.GET("/:id", middleware.RolePermissionMiddleware("user", "get", sqlDB), func(c *gin.Context) { controllers.GetUserByID(c, sqlDB) })
			userRoutes.PUT("/:id", middleware.RolePermissionMiddleware("user", "update", sqlDB), func(c *gin.Context) { controllers.UpdateUser(c, sqlDB) })
			userRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("user", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteUser(c, sqlDB) })
			userRoutes.POST("/change-password", middleware.StrictRateLimiterMiddleware(), func(c *gin.Context) { controllers.ChangePassword(c, sqlDB) })
			userRoutes.GET("/my-permissions", func(c *gin.Context) { controllers.GetMyPermissions(c, sqlDB) })
			userRoutes.POST("/check-permission", func(c *gin.Context) { controllers.CheckPermission(c, sqlDB) })
		}

		staffRoutes := auth.Group("/staff")
		staffRoutes.Use(middleware.TenantMiddleware(sqlDB, mongoClient))
		{
			staffRoutes.POST("", middleware.RolePermissionMiddleware("staff", "create", sqlDB), func(c *gin.Context) { controllers.CreateStaff(c, sqlDB, mongoClient) })
			staffRoutes.GET("", middleware.RolePermissionMiddleware("staff", "get", sqlDB), func(c *gin.Context) { controllers.GetAllStaff(c, sqlDB, mongoClient) })
			staffRoutes.GET("/list", middleware.RolePermissionMiddleware("staff", "get", sqlDB), func(c *gin.Context) { controllers.GetAllStaffList(c, sqlDB, mongoClient) })
			staffRoutes.GET("/:id", middleware.RolePermissionMiddleware("staff", "get", sqlDB), func(c *gin.Context) { controllers.GetStaffByID(c, sqlDB, mongoClient) })
			staffRoutes.PUT("/:id", middleware.RolePermissionMiddleware("staff", "update", sqlDB), func(c *gin.Context) { controllers.UpdateStaff(c, sqlDB, mongoClient) })
			staffRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("staff", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteStaff(c, mongoClient) })
		}

		moduleRoutes := auth.Group("/modules")
		{
			moduleRoutes.POST("", middleware.RolePermissionMiddleware("role", "create", sqlDB), func(c *gin.Context) { controllers.CreateModule(c, sqlDB) })
			moduleRoutes.GET("", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetAllModules(c, sqlDB) })
			moduleRoutes.GET("/:id", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetModuleByID(c, sqlDB) })
			moduleRoutes.PUT("/:id", middleware.RolePermissionMiddleware("role", "update", sqlDB), func(c *gin.Context) { controllers.UpdateModule(c, sqlDB) })
			moduleRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("role", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteModule(c, sqlDB) })
			moduleRoutes.GET("/:id/permissions", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetModulePermissions(c, sqlDB) })
		}
		permissionRoutes := auth.Group("/permissions")
		{
			permissionRoutes.POST("", middleware.RolePermissionMiddleware("role", "create", sqlDB), func(c *gin.Context) { controllers.CreatePermission(c, sqlDB) })
			permissionRoutes.GET("", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetAllPermissions(c, sqlDB) })
			permissionRoutes.GET("/:id", middleware.RolePermissionMiddleware("role", "get", sqlDB), func(c *gin.Context) { controllers.GetPermissionByID(c, sqlDB) })
			permissionRoutes.PUT("/:id", middleware.RolePermissionMiddleware("role", "update", sqlDB), func(c *gin.Context) { controllers.UpdatePermission(c, sqlDB) })
			permissionRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("role", "delete", sqlDB), func(c *gin.Context) { controllers.DeletePermission(c, sqlDB) })
		}
		departmentRoutes := auth.Group("/departments")
		{
			departmentRoutes.POST("", middleware.RolePermissionMiddleware("department", "create", sqlDB), func(c *gin.Context) { controllers.CreateDepartment(c, mongoClient) })
			departmentRoutes.GET("", middleware.RolePermissionMiddleware("department", "get", sqlDB), func(c *gin.Context) { controllers.GetAllDepartments(c, mongoClient) })
			departmentRoutes.GET("/list", middleware.RolePermissionMiddleware("department", "get", sqlDB), func(c *gin.Context) { controllers.GetAllDepartmentsList(c, mongoClient) })
			departmentRoutes.GET("/:id", middleware.RolePermissionMiddleware("department", "get", sqlDB), func(c *gin.Context) { controllers.GetDepartmentByID(c, mongoClient) })
			departmentRoutes.PUT("/:id", middleware.RolePermissionMiddleware("department", "update", sqlDB), func(c *gin.Context) { controllers.UpdateDepartment(c, mongoClient) })
			departmentRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("department", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteDepartment(c, mongoClient) })
			departmentRoutes.POST("/seed", middleware.RolePermissionMiddleware("department", "create", sqlDB), func(c *gin.Context) { controllers.SeedDepartments(c, mongoClient) })
		}

		specialityRoutes := auth.Group("/specialities")
		{
			specialityRoutes.POST("", middleware.RolePermissionMiddleware("speciality", "create", sqlDB), func(c *gin.Context) { controllers.CreateSpeciality(c, mongoClient) })
			specialityRoutes.GET("", middleware.RolePermissionMiddleware("speciality", "get", sqlDB), func(c *gin.Context) { controllers.GetAllSpecialities(c, mongoClient) })
			specialityRoutes.GET("/list", middleware.RolePermissionMiddleware("speciality", "get", sqlDB), func(c *gin.Context) { controllers.GetAllSpecialitiesList(c, mongoClient) })
			specialityRoutes.GET("/department/:department_id", middleware.RolePermissionMiddleware("speciality", "get", sqlDB), func(c *gin.Context) { controllers.GetSpecialitiesByDepartment(c, mongoClient) })
			specialityRoutes.GET("/:id", middleware.RolePermissionMiddleware("speciality", "get", sqlDB), func(c *gin.Context) { controllers.GetSpecialityByID(c, mongoClient) })
			specialityRoutes.PUT("/:id", middleware.RolePermissionMiddleware("speciality", "update", sqlDB), func(c *gin.Context) { controllers.UpdateSpeciality(c, mongoClient) })
			specialityRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("speciality", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteSpeciality(c, mongoClient) })
		}

		shiftRoutes := auth.Group("/shifts")
		{
			shiftRoutes.POST("", middleware.RolePermissionMiddleware("shift", "create", sqlDB), func(c *gin.Context) { controllers.CreateShift(c, mongoClient) })
			shiftRoutes.GET("", middleware.RolePermissionMiddleware("shift", "get", sqlDB), func(c *gin.Context) { controllers.GetAllShifts(c, mongoClient) })
			shiftRoutes.GET("/list", middleware.RolePermissionMiddleware("shift", "get", sqlDB), func(c *gin.Context) { controllers.GetAllShiftsList(c, mongoClient) })
			shiftRoutes.GET("/:id", middleware.RolePermissionMiddleware("shift", "get", sqlDB), func(c *gin.Context) { controllers.GetShiftByID(c, mongoClient) })
			shiftRoutes.PUT("/:id", middleware.RolePermissionMiddleware("shift", "update", sqlDB), func(c *gin.Context) { controllers.UpdateShift(c, mongoClient) })
			shiftRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("shift", "delete", sqlDB), func(c *gin.Context) { controllers.DeleteShift(c, mongoClient) })
		}

		patientRoutes := auth.Group("/patients")
		{
			patientRoutes.POST("", middleware.RolePermissionMiddleware("patient", "create", sqlDB), func(c *gin.Context) { controllers.CreatePatient(c, mongoClient) })
			patientRoutes.GET("", middleware.RolePermissionMiddleware("patient", "get", sqlDB), func(c *gin.Context) { controllers.GetAllPatients(c, mongoClient) })
			patientRoutes.GET("/:id", middleware.RolePermissionMiddleware("patient", "get", sqlDB), func(c *gin.Context) { controllers.GetPatientByID(c, mongoClient) })
			patientRoutes.PUT("/:id", middleware.RolePermissionMiddleware("patient", "update", sqlDB), func(c *gin.Context) { controllers.UpdatePatient(c, mongoClient) })
			patientRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("patient", "delete", sqlDB), func(c *gin.Context) { controllers.DeletePatient(c, mongoClient) })
		}

		fileRoutes := auth.Group("/files")
		{
			fileRoutes.POST("/upload", func(c *gin.Context) { controllers.UploadFile(c, sqlDB) })
			fileRoutes.POST("/presigned-url", controllers.GetPresignedURL)
			fileRoutes.DELETE("/delete", controllers.DeleteFile)
		}
	}
	return r

}
